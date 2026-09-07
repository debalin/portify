import { ArrowLeft, FileText, ExternalLink, AlertTriangle } from 'lucide-react'

interface TermsOfServiceProps {
  onBack: () => void
}

export function TermsOfService({ onBack }: TermsOfServiceProps) {
  const contactEmail = import.meta.env.VITE_CONTACT_EMAIL || 'debalin@debalin.dev'

  return (
    <div className="legal-container">
      <button className="back-btn" onClick={onBack}>
        <ArrowLeft size={18} /> Back to Converter
      </button>

      <div className="legal-card">
        <h1 className="legal-title">Terms of Service</h1>
        <p className="legal-updated">Last Updated: September 6, 2026</p>

        <section className="legal-section">
          <h2>1. Acceptance of Terms</h2>
          <p>
            By accessing or using Portify (&quot;the Service&quot;), you agree to be bound by these Terms of Service (&quot;Terms&quot;). If you do not agree to these Terms, please do not use the Service.
          </p>
        </section>

        <section className="legal-section highlight-box">
          <div className="section-header-icon">
            <ExternalLink className="legal-icon" />
            <h2>2. Third-Party Terms of Service (YouTube, Spotify, Tidal)</h2>
          </div>
          <p>
            Portify connects with third-party digital music and video services. By using Portify to interact with YouTube, Spotify, or Tidal, you agree to comply with and be bound by their respective Terms of Service:
          </p>
          <ul>
            <li>
              <strong>YouTube Terms of Service:</strong> By using Portify to convert or access YouTube playlists, you agree to be bound by the YouTube Terms of Service located at{' '}
              <a href="https://www.youtube.com/t/terms" target="_blank" rel="noopener noreferrer">
                https://www.youtube.com/t/terms
              </a>.
            </li>
            <li>
              <strong>Google Privacy Policy:</strong>{' '}
              <a href="https://www.google.com/policies/privacy" target="_blank" rel="noopener noreferrer">
                https://www.google.com/policies/privacy
              </a>
            </li>
            <li>
              <strong>Spotify Terms of Service:</strong>{' '}
              <a href="https://www.spotify.com/legal/end-user-agreement/" target="_blank" rel="noopener noreferrer">
                https://www.spotify.com/legal/end-user-agreement/
              </a>
            </li>
            <li>
              <strong>Tidal Terms of Use:</strong>{' '}
              <a href="https://tidal.com/terms" target="_blank" rel="noopener noreferrer">
                https://tidal.com/terms
              </a>
            </li>
          </ul>
        </section>

        <section className="legal-section">
          <div className="section-header-icon">
            <FileText className="legal-icon" />
            <h2>3. Permitted Use & Account Conduct</h2>
          </div>
          <p>
            Portify is provided for personal, non-commercial playlist migration and management. You agree that you will not:
          </p>
          <ul>
            <li>Use the Service in any manner that violates applicable laws or third-party intellectual property rights.</li>
            <li>Attempt to bypass, disable, or circumvent API rate limits, daily quotas, or security measures of Portify or third-party providers.</li>
            <li>Scrape, duplicate, or harvest unauthorized content from streaming services.</li>
          </ul>
        </section>

        <section className="legal-section">
          <div className="section-header-icon">
            <AlertTriangle className="legal-icon" />
            <h2>4. Disclaimer of Warranties & Limitation of Liability</h2>
          </div>
          <p>
            Portify is open-source software provided &quot;AS IS&quot; and &quot;AS AVAILABLE&quot;, without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose, and noninfringement.
          </p>
          <p>
            In no event shall the authors or copyright holders be liable for any claim, damages, or other liability arising from the use of the Service or third-party API availability.
          </p>
        </section>

        <section className="legal-section">
          <h2>5. Changes to Terms</h2>
          <p>
            We reserve the right to modify these Terms at any time. Any changes will be posted on this page with an updated &quot;Last Updated&quot; date. Continued use of the Service constitutes your acceptance of the revised Terms.
          </p>
        </section>

        <section className="legal-section">
          <h2>6. Contact</h2>
          <p>
            If you have any questions concerning these Terms of Service, you may contact us at{' '}
            <a href={`mailto:${contactEmail}`}>{contactEmail}</a>.
          </p>
        </section>
      </div>
    </div>
  )
}
