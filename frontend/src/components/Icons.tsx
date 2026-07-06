

export const DefaultMusicIcon = ({ className, title }: { className?: string, title?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    {title && <title>{title}</title>}
    <path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>
  </svg>
)

export const YouTubeMusicIcon = ({ className, title }: { className?: string, title?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor" className={className}>
    {title && <title>{title}</title>}
    <path d="M12 0C5.376 0 0 5.376 0 12s5.376 12 12 12 12-5.376 12-12S18.624 0 12 0zm0 19.104c-3.924 0-7.104-3.18-7.104-7.104S8.076 4.896 12 4.896s7.104 3.18 7.104 7.104-3.18 7.104-7.104 7.104zm0-13.332c-3.432 0-6.228 2.796-6.228 6.228S8.568 18.228 12 18.228s6.228-2.796 6.228-6.228S15.432 5.772 12 5.772zM9.684 15.54V8.46L15.816 12l-6.132 3.54z"/>
  </svg>
)

export const SpotifyIcon = ({ className, title }: { className?: string, title?: string }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="1em" height="1em" 
    viewBox="0 0 24 24" 
    fill="currentColor" 
    className={className}
  >
    {title && <title>{title}</title>}
    <path d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.6 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.54.659.3 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.6.18-1.2.72-1.38 4.2-1.26 11.28-1.02 15.72 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z" />
  </svg>
)

export const TidalIcon = ({ className, title }: { className?: string, title?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor" className={className}>
    {title && <title>{title}</title>}
    <path d="M12.01 5.334L8.683 2 5.341 5.334l3.327 3.322 3.342-3.322zM8.683 8.656L5.341 12l3.342 3.333 3.327-3.333-3.327-3.344zM12.01 12l-3.327 3.333L12.01 18.667l3.342-3.334L12.01 12zM15.353 8.656L12.01 12l3.342 3.333L18.678 12l-3.325-3.344zM18.678 5.334L15.353 2 12.01 5.334l3.343 3.322 3.325-3.322z"/>
  </svg>
)

export const DynamicProviderIcon = ({ providerId, className }: { providerId: string, className?: string }) => {
  if (providerId === 'spotify') return <SpotifyIcon className={className} />
  if (providerId === 'youtube') return <YouTubeMusicIcon className={className} />
  if (providerId === 'tidal') return <TidalIcon className={className} />
  return <DefaultMusicIcon className={className} />
}
