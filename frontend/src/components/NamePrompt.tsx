import { useState } from 'react';
import '../styles/NamePrompt.css';

interface NamePromptProps {
  onSubmit: (name: string) => void | Promise<void>;
  onCancel?: () => void;
}

export function NamePrompt({ onSubmit }: NamePromptProps) {
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const validateName = (value: string): string => {
    if (!value || value.length === 0) {
      return 'Name must be 1-20 characters';
    }

    if (value.length > 20) {
      return 'Name must be 1-20 characters';
    }

    // Validate: alphanumeric, spaces, dashes, underscores only
    const validPattern = /^[a-zA-Z0-9\s\-_]+$/;
    if (!validPattern.test(value)) {
      return 'Name can only contain alphanumeric characters, spaces, dashes, or underscores';
    }

    return '';
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const validationError = validateName(name);
    if (validationError) {
      setError(validationError);
      return;
    }

    setError('');
    setIsSubmitting(true);

    try {
      await onSubmit(name);
    } catch {
      setError('Failed to submit name. Please try again.');
      setIsSubmitting(false);
    }
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setName(value);

    // Clear error on valid input
    if (!value || validateName(value) === '') {
      setError('');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !isSubmitting) {
      handleSubmit(e as unknown as React.FormEvent<HTMLFormElement>);
    }
  };

  return (
    <div className="name-prompt-overlay">
      <div className="name-prompt-modal">
        <h1>Enter Your Name</h1>
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            placeholder="Your name"
            value={name}
            onChange={handleNameChange}
            onKeyDown={handleKeyDown}
            disabled={isSubmitting}
            maxLength={20}
          />
          {error && <p className="name-prompt-error">{error}</p>}
          <button type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Joining...' : 'Join Game'}
          </button>
        </form>
      </div>
    </div>
  );
}
