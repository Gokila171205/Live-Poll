export function Toggle({
  id,
  checked,
  onChange,
  label,
  description,
  disabled = false,
  className = '',
}) {
  return (
    <label
      htmlFor={id}
      className={`pollflow-toggle-wrapper ${disabled ? 'disabled' : ''} ${className}`}
    >
      <div className="toggle-text-block">
        {label && <span className="toggle-label">{label}</span>}
        {description && <span className="toggle-description">{description}</span>}
      </div>
      <button
        type="button"
        id={id}
        role="switch"
        aria-checked={checked}
        disabled={disabled}
        onClick={() => !disabled && onChange(!checked)}
        className={`toggle-switch-btn ${checked ? 'checked' : ''}`}
      >
        <span className="toggle-switch-thumb" />
      </button>
    </label>
  )
}
