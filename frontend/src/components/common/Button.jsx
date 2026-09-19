export function Button({
  children,
  variant = 'primary', // 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline'
  size = 'md', // 'sm' | 'md' | 'lg'
  isLoading = false,
  disabled = false,
  className = '',
  type = 'button',
  icon = null,
  onClick,
  ...props
}) {
  const baseClass = 'custom-btn'
  const variantClass = `btn-${variant}`
  const sizeClass = `btn-${size}`
  const loadingClass = isLoading ? 'btn-loading' : ''

  return (
    <button
      type={type}
      disabled={disabled || isLoading}
      className={`${baseClass} ${variantClass} ${sizeClass} ${loadingClass} ${className}`.trim()}
      onClick={onClick}
      {...props}
    >
      {isLoading ? (
        <span className="spinner-inline"></span>
      ) : icon ? (
        <span className="btn-icon">{icon}</span>
      ) : null}
      <span className="btn-text">{children}</span>
    </button>
  )
}
