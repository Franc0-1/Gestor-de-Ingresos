export function EmptyState({ title, message }: { title: string; message: string }) {
  return <div className="empty-state"><span className="empty-mark" aria-hidden="true"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="4" y="4" width="16" height="16" rx="4" /><path d="M8 9h8m-8 4h5" /></svg></span><strong>{title}</strong><p>{message}</p></div>;
}
