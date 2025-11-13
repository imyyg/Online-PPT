export function getErrorMessage(error) {
  if (error.response?.data?.message) {
    return error.response.data.message
  }
  
  if (error.response?.data?.code) {
    return getErrorMessageByCode(error.response.data.code)
  }
  
  if (error.message) {
    return error.message
  }
  
  return 'An unexpected error occurred'
}

export function getErrorMessageByCode(code) {
  const messages = {
    invalid_request: 'Invalid request. Please check your input.',
    invalid_captcha: 'Invalid or expired captcha. Please try again.',
    invalid_code: 'Invalid or expired verification code.',
    email_exists: 'This email is already registered.',
    invalid_credentials: 'Invalid email or password.',
    session_not_found: 'Session expired. Please login again.',
    rate_limited: 'Too many requests. Please try again later.',
    too_many_attempts: 'Too many attempts. Please request a new code.',
    server_error: 'Server error. Please try again later.'
  }
  
  return messages[code] || 'An error occurred'
}

export function isNetworkError(error) {
  return !error.response && error.request
}

export function isAuthError(error) {
  return error.response?.status === 401
}

export function isValidationError(error) {
  return error.response?.status === 400
}
