package app

import (
	"fmt"
	"log"
	"time"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/pocket-ledger/backend/internal/application/auth"
	"github.com/katalabut/pocket-ledger/backend/internal/application/budget"
	"github.com/katalabut/pocket-ledger/backend/internal/application/fx"
	"github.com/katalabut/pocket-ledger/backend/internal/application/importer"
	"github.com/katalabut/pocket-ledger/backend/internal/application/ledger"
	"github.com/katalabut/pocket-ledger/backend/internal/application/reports"
	"github.com/katalabut/pocket-ledger/backend/internal/infrastructure/ecb"
	"github.com/katalabut/pocket-ledger/backend/internal/infrastructure/email"
	"github.com/katalabut/pocket-ledger/backend/internal/infrastructure/sqliterepo"
	"github.com/katalabut/pocket-ledger/backend/internal/interfaces/httpapi"
	"github.com/katalabut/pocket-ledger/backend/internal/interfaces/httpauth"
	"github.com/katalabut/pocket-ledger/backend/internal/interfaces/httpserver"
	"github.com/katalabut/pocket-ledger/backend/internal/platform/appconfig"
	"github.com/katalabut/pocket-ledger/backend/internal/platform/migrator"
	"github.com/katalabut/pocket-ledger/backend/internal/platform/sqlite"
)

func New() (*fastapp.App, error) {
	cfg, err := appconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if cfg.Auth.JWTSecret == "" {
		return nil, fmt.Errorf("AUTH_JWT_SECRET is required")
	}
	emailTestMode := cfg.Email.Mode == "log" || cfg.SMTP.TestMode
	if !emailTestMode && (cfg.SMTP.Host == "" || cfg.SMTP.Username == "" || cfg.SMTP.Password == "" || cfg.SMTP.From == "") {
		return nil, fmt.Errorf("SMTP config is required (HOST/USERNAME/PASSWORD/FROM) unless EMAIL_MODE=log or SMTP_TEST_MODE=true")
	}

	sqliteSvc, err := sqlite.New(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("init sqlite: %w", err)
	}
	migSvc := migrator.New(sqliteSvc.DB(), "migrations/sqlite")

	// Enable FK enforcement per connection
	if _, err := sqliteSvc.DB().Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	var sender auth.EmailSender
	if emailTestMode {
		log.Println("email test mode enabled (login codes will be logged and not sent)")
		sender = email.NewLoggingSender(log.Default(), cfg.SMTP.From)
	} else {
		sender = email.NewSMTPSender(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.From)
	}
	codesRepo := sqliterepo.NewAuthCodeRepo(sqliteSvc.DB())
	usersRepo := sqliterepo.NewUserRepo(sqliteSvc.DB())
	jwtIssuer := httpauth.NewIssuer(cfg.Auth.JWTSecret, cfg.Auth.Issuer)
	authSvc := auth.NewService(codesRepo, usersRepo, sender, jwtIssuer, time.Duration(cfg.Auth.CodeTTLSeconds)*time.Second, cfg.Auth.CodeLengthDigits, nil)

	accountRepo := sqliterepo.NewAccountRepo(sqliteSvc.DB())
	categoryRepo := sqliterepo.NewCategoryRepo(sqliteSvc.DB())
	transactionRepo := sqliterepo.NewTransactionRepo(sqliteSvc.DB())
	splitRepo := sqliterepo.NewSplitRepo(sqliteSvc.DB())
	ledgerSvc := ledger.NewService(accountRepo, categoryRepo, transactionRepo, splitRepo, nil)

	importProfileRepo := sqliterepo.NewImportProfileRepo(sqliteSvc.DB())
	importRepo := sqliterepo.NewImportRepo(sqliteSvc.DB())
	importSvc := importer.NewService(importProfileRepo, importRepo, transactionRepo, nil)

	fxRateRepo := sqliterepo.NewFXRateRepo(sqliteSvc.DB())
	ecbClient := ecb.NewClient()
	fxSvc := fx.NewService(fxRateRepo, ecbClient, cfg.FX.BaseCurrency, nil)
	reportsSvc := reports.NewService(ledgerSvc, fxSvc)

	budgetRepo := sqliterepo.NewBudgetRepo(sqliteSvc.DB())
	budgetSvc := budget.NewService(budgetRepo, ledgerSvc, fxSvc, nil)

	api := httpapi.New(authSvc, ledgerSvc, importSvc, fxSvc, reportsSvc, budgetSvc, httpapi.Config{JWTSecret: cfg.Auth.JWTSecret, Issuer: cfg.Auth.Issuer})
	httpSrv := httpserver.New(cfg.HTTP.Addr, api.Handler())

	a := fastapp.New(cfg.App).
		Add(migSvc).
		Add(sqliteSvc).
		Add(httpSrv)

	return a, nil
}
