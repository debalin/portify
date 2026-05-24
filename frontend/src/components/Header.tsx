import { SpotifyIcon, YouTubeMusicIcon, TidalIcon } from './Icons'

export function Header() {
  return (
    <header className="header">
      <h1 className="title">Portify</h1>
      <p className="subtitle">The Universal Playlist Converter</p>
      <div className="supported-providers-banner">
        <span className="banner-text">Supported Platforms:</span>
        <div className="banner-icons">
          <SpotifyIcon className="banner-icon spotify" title="Spotify" />
          <YouTubeMusicIcon className="banner-icon youtube" title="YouTube Music" />
          <TidalIcon className="banner-icon tidal" title="Tidal" />
        </div>
      </div>
    </header>
  )
}
