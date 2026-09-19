export function PollFlowLogo({ className = '', height = 32, showText = true }) {
  return (
    <div
      className={`pollflow-brand-logo ${className}`}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '10px',
        textDecoration: 'none',
        userSelect: 'none',
      }}
    >
      <svg
        width={height}
        height={height}
        viewBox="0 0 32 32"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        style={{ flexShrink: 0, borderRadius: '8px' }}
      >
        <rect width="32" height="32" rx="8" fill="#2563EB" />
        {/* Subtle inner bar illustration representing live polls */}
        <rect x="7" y="15" width="4" height="10" rx="2" fill="white" opacity="0.95" />
        <rect x="14" y="9" width="4" height="16" rx="2" fill="white" />
        <rect x="21" y="13" width="4" height="12" rx="2" fill="white" opacity="0.95" />
      </svg>
      {showText && (
        <span
          style={{
            fontWeight: 700,
            fontSize: `${Math.max(18, Math.round(height * 0.6))}px`,
            letterSpacing: '-0.025em',
            color: 'var(--text-primary, #0F172A)',
            display: 'inline-flex',
            alignItems: 'center',
          }}
        >
          Poll<span style={{ color: '#2563EB' }}>Flow</span>
        </span>
      )}
    </div>
  )
}
