import { ArrowLeft, ShieldCheck, Lock, Trash2, Key } from 'lucide-react'

interface PrivacyPolicyProps {
  onBack: () => void
}

export function PrivacyPolicy({ onBack }: PrivacyPolicyProps) {
  const contactEmail = import.meta.env.VITE_CONTACT_EMAIL || 'debalin@debalin.dev'

  return (
    <div className="legal-container">
      <button className="back-btn" onClick={onBack}>
        <ArrowLeft size={18} /> Back to Converter
      </button>

      <div className="legal-card">
        <h1 className="legal-title">Privacy Policy</h1>
        <p className="legal-updated">Last Updated: September 12, 2026</p>

        <section className="legal-section">
          <h2>1. Overview</h2>
          <p>
            Portify (&quot;we&quot;, &quot;our&quot;, or &quot;the application&quot;) is an open-source playlist conversion tool designed to help users transfer their music playlists across music streaming platforms, including YouTube Music, Spotify, and Tidal. We are committed to protecting your privacy and handling your data with complete transparency.
          </p>
        </section>

        <section className="legal-section highlight-box">
          <div className="section-header-icon">
            <ShieldCheck className="legal-icon" />
            <h2>2. YouTube API Services Compliance & Disclosure</h2>
          </div>
          <p>
            Portify uses <strong>YouTube API Services</strong> to fetch and create playlists on YouTube Music on your behalf.
          </p>
          <p>
            By using Portify to interact with YouTube services, you acknowledge and agree that your usage is governed by the:
          </p>
          <ul>
            <li>
              <strong>YouTube Terms of Service:</strong>{' '}
              <a href="https://www.youtube.com/t/terms" target="_blank" rel="noopener noreferrer">
                https://www.youtube.com/t/terms
              </a>
            </li>
            <li>
              <strong>Google Privacy Policy:</strong>{' '}
              <a href="https://www.google.com/policies/privacy" target="_blank" rel="noopener noreferrer">
                https://www.google.com/policies/privacy
              </a>
            </li>
          </ul>
        </section>

        <section className="legal-section">
          <div className="section-header-icon">
            <Lock className="legal-icon" />
            <h2>3. Information We Collect and Process</h2>
          </div>
          <p>
            Portify processes only the minimal, necessary data required to complete your requested playlist conversions:
          </p>
          <ul>
            <li>
              <strong>Playlist Metadata:</strong> Playlist titles, descriptions, and track listings (track title, artist name, and album name) retrieved from or sent to streaming providers (e.g., YouTube Music, Spotify, Tidal).
            </li>
            <li>
              <strong>OAuth Authorization Tokens:</strong> Ephemeral access tokens obtained directly via standard OAuth 2.0 authorization flows with Google/YouTube, Spotify, and Tidal.
            </li>
          </ul>
          <p>
            Portify <strong>does not</strong> collect, track, or store personal user profile details, email addresses, passwords, viewing history, or payment information.
          </p>
        </section>

        <section className="legal-section">
          <div className="section-header-icon">
            <ShieldCheck className="legal-icon" />
            <h2>4. How We Use, Process, and Share Your Information (Internal & External Parties)</h2>
          </div>
          <p>
            In compliance with YouTube API Services Developer Policies (Policy III.A.2e), we explicitly disclose how user information and API data are used, processed, and shared:
          </p>
          <ul>
            <li>
              <strong>Purpose of Use:</strong> User data and YouTube API data are used <em>exclusively</em> to perform user-requested playlist migrations (reading playlist tracks from a designated source and creating corresponding playlists or adding matching tracks to a designated destination). We do not use API data for any secondary purpose, profiling, advertising, or machine learning training.
            </li>
            <li>
              <strong>Internal Parties:</strong> Portify is an open-source tool. We do not have internal staff, commercial teams, or support agents who view, inspect, or access user data or API data. No internal databases or storage volumes contain user data.
            </li>
            <li>
              <strong>External Parties (No Selling or Third-Party Sharing):</strong> We <strong>do not sell, rent, trade, license, or share</strong> user data or YouTube API Data with any third-party advertisers, data brokers, advertising networks, or external analytics vendors.
            </li>
            <li>
              <strong>Authorized Service Providers Only:</strong> Data is transmitted solely over encrypted HTTPS connections to the official endpoints of the music streaming services you have specifically selected and authenticated with (Google/YouTube Data API v3, Spotify Web API, and Tidal API) strictly to execute the requested conversion actions on your behalf.
            </li>
          </ul>
        </section>

        <section className="legal-section">
          <div className="section-header-icon">
            <Key className="legal-icon" />
            <h2>5. Device Storage, Cookies, and Browser Technologies</h2>
          </div>
          <p>
            In compliance with YouTube API Services Developer Policies (Policy III.A.2g), we disclose how information is stored directly on your device:
          </p>
          <ul>
            <li>
              <strong>Local Browser Storage (<code>sessionStorage</code>):</strong> Portify uses your browser&apos;s client-side <code>sessionStorage</code> directly on your device to maintain your active conversion session. Specifically, we store:
              <ul>
                <li><code>portifyAuthTokens</code>: Temporary OAuth 2.0 access tokens required to authorize requests to streaming providers during your active session.</li>
                <li><code>portifySource</code> &amp; <code>portifyDest</code>: Your selected source and destination provider choices.</li>
                <li><code>portifyPlaylistId</code> &amp; <code>portifyDestPlaylistId</code>: The identifiers of playlists selected for conversion.</li>
              </ul>
            </li>
            <li>
              <strong>Cookies:</strong> Portify itself does not place tracking cookies, marketing cookies, or third-party web beacons on users&apos; devices or browsers. Third-party authentication providers (such as Google, Spotify, or Tidal) may place or recognize their own authentication cookies on their respective domains when you interact with their OAuth login dialogs, governed by their respective privacy policies.
            </li>
            <li>
              <strong>Storage Control:</strong> <code>sessionStorage</code> is strictly sandboxed to your browser tab. Closing the browser tab or clicking &quot;Log out&quot; automatically and permanently purges all stored tokens and session data from your device.
            </li>
          </ul>
        </section>

        <section className="legal-section">
          <div className="section-header-icon">
            <Lock className="legal-icon" />
            <h2>6. API Data Storage, Refresh, and Deletion Lifecycle</h2>
          </div>
          <p>
            In compliance with YouTube API Services Developer Policies (Policy III.E.4a-g), we disclose our data retention and deletion schedule:
          </p>
          <ul>
            <li>
              <strong>Zero Persistent Storage (No Caching):</strong> Portify operates on a completely stateless architecture. We <strong>do not store, cache, or archive</strong> YouTube API Data (including track listings, video metadata, or user profile information) on any persistent server database, filesystem, or permanent cache.
            </li>
            <li>
              <strong>Real-Time Ephemeral Processing:</strong> YouTube API data is fetched on-demand into volatile server memory solely during the active conversion stream. As each track is matched and inserted, progress is streamed to the user via Server-Sent Events (SSE). Once the conversion completes, all in-memory track data is immediately released and discarded.
            </li>
            <li>
              <strong>Retention Period:</strong> 0 seconds on server/persistent storage. All client-side session tokens are cleared upon closing the browser session.
            </li>
          </ul>
        </section>

        <section className="legal-section">
          <div className="section-header-icon">
            <Trash2 className="legal-icon" />
            <h2>7. Revoking Access & Data Deletion</h2>
          </div>
          <p>
            You have full control over Portify&apos;s access to your Google/YouTube account data. You can revoke Portify&apos;s access permissions at any time:
          </p>
          <ul>
            <li>
              Visit the <strong>Google Security Settings / Permissions page:</strong>{' '}
              <a href="https://myaccount.google.com/permissions" target="_blank" rel="noopener noreferrer">
                https://myaccount.google.com/permissions
              </a>
            </li>
            <li>
              Locate Portify under &quot;Third-party apps with account access&quot; and click <strong>Remove Access</strong>.
            </li>
          </ul>
          <p>
            <strong>Data Deletion:</strong> Because Portify does not store personal data or credentials on any server or database, simply ending your browser session or clicking &quot;Log out&quot; deletes all active access tokens from your device. If you have any inquiries regarding data deletion, you may contact the maintainer at{' '}
            <a href={`mailto:${contactEmail}`}>{contactEmail}</a>.
          </p>
        </section>

        <section className="legal-section">
          <h2>8. Contact Information</h2>
          <p>
            If you have questions or concerns about this Privacy Policy or Portify&apos;s privacy practices, please contact us at{' '}
            <a href={`mailto:${contactEmail}`}>{contactEmail}</a> or open an issue on GitHub at{' '}
            <a href="https://github.com/debalin/portify" target="_blank" rel="noopener noreferrer">
              https://github.com/debalin/portify
            </a>.
          </p>
        </section>
      </div>
    </div>
  )
}
