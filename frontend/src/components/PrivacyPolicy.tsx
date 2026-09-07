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
        <p className="legal-updated">Last Updated: September 6, 2026</p>

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
              <strong>Playlist Metadata:</strong> Playlist titles, descriptions, and track listings (track title, artist name, and album name) provided by source platforms (e.g., Spotify, Tidal).
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
            <Key className="legal-icon" />
            <h2>4. Data Storage and Retention</h2>
          </div>
          <p>
            <strong>We do not maintain any persistent databases of user data.</strong>
          </p>
          <ul>
            <li>
              OAuth access tokens are stored <strong>strictly on your local device in browser session storage</strong> (<code>sessionStorage</code>). They are never written to disk or stored in any remote database by Portify.
            </li>
            <li>
              Tokens are transmitted securely over encrypted HTTPS connections directly to official streaming service APIs to perform your conversion.
            </li>
            <li>
              When you close your browser tab or clear your browser data, all session tokens and temporary state are permanently removed.
            </li>
          </ul>
        </section>

        <section className="legal-section">
          <div className="section-header-icon">
            <Trash2 className="legal-icon" />
            <h2>5. Revoking Access & Data Deletion</h2>
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
          <h2>6. Contact Information</h2>
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
