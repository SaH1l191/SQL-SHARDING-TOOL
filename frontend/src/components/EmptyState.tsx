interface Props {
  icon?: string
  title: string
  description?: string
  action?: { label: string; onClick: () => void }
}

export function EmptyState({ icon = '⬡', title, description, action }: Props) {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      <span className="text-4xl text-gray-700 mb-3">{icon}</span>
      <p className="text-sm font-medium text-gray-400">{title}</p>
      {description && <p className="mt-1 text-xs text-gray-600 max-w-xs">{description}</p>}
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-700 transition-colors"
        >
          {action.label}
        </button>
      )}
    </div>
  )
}
