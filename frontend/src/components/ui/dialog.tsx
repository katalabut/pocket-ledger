import type { ReactNode } from 'react'
import { Button } from './button'
import { Card, CardContent, CardHeader, CardTitle } from './card'

export function Dialog({ open, onClose, title, children, footer, width = 'max-w-2xl' }: { open: boolean; onClose: () => void; title: string; children: ReactNode; footer?: ReactNode; width?: string }) {
  if (!open) return null
  return (
    <div className="fixed inset-0 z-50 bg-black/35 p-4" onClick={onClose}>
      <div className={`mx-auto mt-10 ${width}`} onClick={e => e.stopPropagation()}>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between border-b border-slate-200">
            <CardTitle>{title}</CardTitle>
            <Button variant="ghost" size="sm" onClick={onClose}>Close</Button>
          </CardHeader>
          <CardContent className="pt-4">{children}{footer ? <div className="mt-4 flex gap-2">{footer}</div> : null}</CardContent>
        </Card>
      </div>
    </div>
  )
}
