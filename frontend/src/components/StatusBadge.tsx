interface Props { status: string }

export function StatusBadge({ status }: Props) {
  const styles: Record<string, string> = {
    active:   'bg-emerald-900/40 text-emerald-400 ring-emerald-800/50',
    inactive: 'bg-gray-800 text-gray-500 ring-gray-700',
    pending:  'bg-amber-900/40 text-amber-400 ring-amber-800/50',
    error:    'bg-red-900/40 text-red-400 ring-red-800/50',
  }
  const cls = styles[status] ?? styles.inactive
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium ring-1 ${cls}`}>
      {status}
    </span>
  )
}
