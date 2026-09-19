export function Input({
  label,
  error,
  icon = null,
  suffix = null,
  id,
  type = 'text',
  className = '',
  required = false,
  helperText,
  ...props
}) {
  const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`

  return (
    <div className={`form-field-group ${error ? 'has-error' : ''} ${className}`.trim()}>
      {label && (
        <label htmlFor={inputId} className="form-label">
          {label} {required && <span className="text-required">*</span>}
        </label>
      )}
      <div className="input-wrapper">
        {icon && <span className="input-icon-prefix">{icon}</span>}
        <input
          id={inputId}
          type={type}
          required={required}
          className={`custom-input ${icon ? 'with-prefix' : ''} ${suffix ? 'with-suffix' : ''}`}
          {...props}
        />
        {suffix && <span className="input-suffix-wrapper">{suffix}</span>}
      </div>
      {error ? (
        <p className="input-error-msg">{error}</p>
      ) : helperText ? (
        <p className="input-helper-msg">{helperText}</p>
      ) : null}
    </div>
  )
}
