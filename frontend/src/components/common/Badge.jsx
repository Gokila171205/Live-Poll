export function Badge({
  children,
  variant = 'default', // 'live' | 'active' | 'closed' | 'indigo' | 'emerald' | 'rose' | 'amber'
  isPulse = false,
  size = 'md',
  className = '',
}) {
  const isLive = variant === 'live' || isPulse

  return (
    <span className={`badge badge-${variant} badge-${size} ${className}`.trim()}>
      {isLive && <span className="pulse-dot"></span>}
      <span className="badge-text">{children}</span>
    </span>
  )
}
