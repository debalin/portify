interface FooterProps {
  onNavigate: (path: string) => void
}

export function Footer({ onNavigate }: FooterProps) {
  return (
    <footer className="app-footer">
      <div className="footer-attribution">
        <span className="footer-badge">
          Powered by <strong>YouTube API Services</strong>
        </span>
      </div>

      <div className="footer-links">
        <button 
          className="footer-link-btn"
          onClick={() => onNavigate('/privacy')}
        >
          Privacy Policy
        </button>
        <span className="footer-divider">•</span>
        <button 
          className="footer-link-btn"
          onClick={() => onNavigate('/terms')}
        >
          Terms of Service
        </button>
        <span className="footer-divider">•</span>
        <a 
          href="https://www.google.com/policies/privacy" 
          target="_blank" 
          rel="noopener noreferrer" 
          className="footer-external-link"
        >
          Google Privacy Policy
        </a>
        <span className="footer-divider">•</span>
        <a 
          href="https://www.youtube.com/t/terms" 
          target="_blank" 
          rel="noopener noreferrer" 
          className="footer-external-link"
        >
          YouTube Terms of Service
        </a>
      </div>

      <div className="footer-copyright">
        &copy; {new Date().getFullYear()} Portify. Open source under MIT License.
      </div>
    </footer>
  )
}
