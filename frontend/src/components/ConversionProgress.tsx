interface ConversionProgressProps {
  progress: {
    status: number
    message: string
    converted: number
    total: number
  } | null
}

export function ConversionProgress({ progress }: ConversionProgressProps) {
  if (!progress || progress.status <= 0 || progress.status >= 3) {
    return null
  }

  const showPercentage = progress.total > 0
  const percentage = showPercentage ? Math.round((progress.converted / progress.total) * 100) : 0

  return (
    <div className="progress-container">
      <div className="progress-header">
        <span className="progress-status">{progress.message}</span>
        {showPercentage && (
          <span className="progress-count">{percentage}%</span>
        )}
      </div>
      {showPercentage && (
        <div className="progress-bar-bg">
          <div 
            className="progress-bar-fill" 
            style={{ width: `${percentage}%` }}
          ></div>
        </div>
      )}
    </div>
  )
}
