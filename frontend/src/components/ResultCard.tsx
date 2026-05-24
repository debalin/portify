import { CheckCircle, AlertCircle } from 'lucide-react'

interface ResultCardProps {
  result: {
    success: boolean
    message: string
    url?: string
    failedTracks?: any[]
  } | null
  progress: {
    converted: number
    total: number
  } | null
  destPlaylistId: string
}

export function ResultCard({ result, progress, destPlaylistId }: ResultCardProps) {
  if (!result) {
    return null
  }

  const showStats = result.success && progress && progress.total > 0

  return (
    <div className={`result-card ${result.success ? 'success' : 'error'}`}>
      {result.success ? (
        <CheckCircle className="result-icon" />
      ) : (
        <AlertCircle className="result-icon error" />
      )}
      <div>
        <p className="result-msg">{result.message}</p>
        {showStats && (
          <p className="result-stats">Converted {progress.converted} of {progress.total} tracks.</p>
        )}
        {result.url && (
          <a href={result.url} target="portify_dest" rel="noreferrer" className="result-link">
            {destPlaylistId ? 'Open Playlist' : 'Open New Playlist'} &rarr;
          </a>
        )}
        {result.failedTracks && result.failedTracks.length > 0 && (
          <details style={{
            marginTop: '12px',
            fontSize: '13px',
            background: 'rgba(255,100,100,0.1)',
            borderRadius: '8px',
            padding: '8px',
            border: '1px solid rgba(255,100,100,0.2)'
          }}>
            <summary style={{ cursor: 'pointer', color: '#ff6b6b', fontWeight: 600 }}>
              Failed to compile {result.failedTracks.length} tracks
            </summary>
            <ul style={{
              marginTop: '8px',
              paddingLeft: '20px',
              maxHeight: '150px',
              overflowY: 'auto',
              color: '#ffaaaa'
            }}>
              {result.failedTracks.map((t, idx) => (
                <li key={idx} style={{ marginBottom: '4px' }}>
                  {t.title} - {t.artist}
                </li>
              ))}
            </ul>
          </details>
        )}
      </div>
    </div>
  )
}
