const BASE = '';

function token(): string | null {
  return localStorage.getItem('pl_token');
}

export function setToken(t: string) {
  localStorage.setItem('pl_token', t);
}

export function clearToken() {
  localStorage.removeItem('pl_token');
}

export function hasToken(): boolean {
  return !!token();
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  const tok = token();
  if (tok) headers['Authorization'] = `Bearer ${tok}`;
  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });
  if (res.status === 204) return undefined as T;
  if (res.status === 401) {
    clearToken();
    window.location.reload();
    throw new Error('unauthorized');
  }
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data as T;
}

// Auth
export const authRequestCode = (email: string) => request<{ sent: boolean }>('POST', '/auth/request_code', { email });
export const authConfirmCode = (email: string, code: string) => request<{ token: string }>('POST', '/auth/confirm_code', { email, code });

// Accounts
export interface Account {
  ID: string;
  Name: string;
  Currency: string;
  Type: string;
  InitialBalanceMinor: number;
  CreatedAt: string;
  UpdatedAt: string;
}
export const listAccounts = () => request<Account[]>('GET', '/api/accounts');
export const createAccount = (data: { name: string; currency: string; type: string; initial_balance_minor?: number }) => request<Account>('POST', '/api/accounts', data);
export const updateAccount = (id: string, data: Partial<{ name: string; currency: string; type: string; initial_balance_minor: number }>) => request<Account>('PATCH', `/api/accounts/${id}`, data);
export const deleteAccount = (id: string) => request<void>('DELETE', `/api/accounts/${id}`);

// Categories
export interface Category {
  ID: string;
  Name: string;
  ParentID: string | null;
  CreatedAt: string;
  UpdatedAt: string;
}
export const listCategories = () => request<Category[]>('GET', '/api/categories');
export const createCategory = (data: { name: string; parent_id?: string }) => request<Category>('POST', '/api/categories', data);
export const updateCategory = (id: string, data: Partial<{ name: string; parent_id: string }>) => request<Category>('PATCH', `/api/categories/${id}`, data);
export const deleteCategory = (id: string) => request<void>('DELETE', `/api/categories/${id}`);

// Transactions
export interface Transaction {
  ID: string;
  AccountID: string;
  OccurredAt: string;
  AmountMinor: number;
  Currency: string;
  Description: string;
  CategoryID: string | null;
  DedupeKey: string | null;
  CreatedAt: string;
  UpdatedAt: string;
}
export interface TransactionList {
  items: Transaction[];
  total: number;
}
export const listTransactions = (params?: Record<string, string>) => {
  const qs = params ? '?' + new URLSearchParams(params).toString() : '';
  return request<TransactionList>('GET', `/api/transactions${qs}`);
};
export const createTransaction = (data: { account_id: string; occurred_at: string; amount_minor: number; currency: string; description: string; category_id?: string }) =>
  request<Transaction>('POST', '/api/transactions', data);
export const updateTransaction = (id: string, data: Partial<{ occurred_at: string; amount_minor: number; currency: string; description: string; category_id: string; account_id: string }>) =>
  request<Transaction>('PATCH', `/api/transactions/${id}`, data);
export const deleteTransaction = (id: string) => request<void>('DELETE', `/api/transactions/${id}`);

// Splits
export interface Split {
  ID: string;
  TransactionID: string;
  CategoryID: string;
  AmountMinor: number;
  CreatedAt: string;
}
export const getSplits = (txID: string) => request<Split[]>('GET', `/api/transactions/${txID}/splits`);
export const replaceSplits = (txID: string, splits: { category_id: string; amount_minor: number }[]) =>
  request<Split[]>('POST', `/api/transactions/${txID}/splits`, splits);

// Import profiles
export interface ImportProfile {
  ID: string;
  Name: string;
  AccountID: string;
  Separator: string;
  DateFormat: string;
  ColumnMapping: Record<string, number>;
  AmountSignFlip: boolean;
  SkipHeaderRows: number;
}
export const listImportProfiles = () => request<ImportProfile[]>('GET', '/api/import-profiles');
export const createImportProfile = (data: { name: string; account_id: string; separator: string; date_format: string; column_mapping: Record<string, number>; amount_sign_flip: boolean; skip_header_rows: number }) =>
  request<ImportProfile>('POST', '/api/import-profiles', data);

// Imports
export interface ImportSession {
  ID: string;
  ProfileID: string;
  AccountID: string;
  Filename: string;
  Status: string;
  TotalRows: number;
  ImportedRows: number;
  SkippedRows: number;
  ErrorRows: number;
}
export interface ImportRow {
  ID: string;
  RowNumber: number;
  RawData: string;
  Status: string;
  ErrorMessage: string | null;
}
export interface ImportPreview {
  import: ImportSession;
  rows: ImportRow[];
}

export const uploadImport = async (profileID: string, file: File): Promise<ImportSession> => {
  const form = new FormData();
  form.append('file', file);
  form.append('profile_id', profileID);
  const headers: Record<string, string> = {};
  const tok = token();
  if (tok) headers['Authorization'] = `Bearer ${tok}`;
  const res = await fetch(`${BASE}/api/imports/upload`, { method: 'POST', headers, body: form });
  if (!res.ok) { const d = await res.json(); throw new Error(d.error || res.statusText); }
  return res.json();
};
export const previewImport = (id: string) => request<ImportPreview>('GET', `/api/imports/${id}/preview`);
export const commitImport = (id: string) => request<ImportSession>('POST', `/api/imports/${id}/commit`);

// Reports
export interface SpendingRow { CategoryID: string; CategoryName: string; TotalMinor: number; }
export interface AccountBalance { AccountID: string; AccountName: string; Currency: string; BalanceMinor: number; BalanceBaseMinor: number; }
export const reportSpending = (params: Record<string, string>) => {
  const qs = '?' + new URLSearchParams(params).toString();
  return request<SpendingRow[]>('GET', `/api/reports/spending${qs}`);
};
export const reportBalances = () => request<AccountBalance[]>('GET', '/api/reports/balances');

// Budgets
export interface Budget {
  ID: string;
  Month: string;
  CategoryID: string;
  CategoryName?: string;
  PlannedMinor: number;
  SpentMinor?: number;
  RemainingMinor?: number;
}
export const listBudgets = (month: string) => request<Budget[]>('GET', `/api/budgets?month=${month}`);
export const upsertBudget = (data: { month: string; category_id: string; planned_minor: number }) => request<Budget>('POST', '/api/budgets', data);

// FX
export const fetchFXRates = (date?: string) => {
  const qs = date ? `?date=${date}` : '';
  return request<Record<string, number>>('GET', `/api/fx/rates${qs}`);
};
export const syncFXRates = () => request<{ synced: number }>('POST', '/api/fx/sync');
