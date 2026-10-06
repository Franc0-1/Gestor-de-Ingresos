export function ErrorState({ message, retry }: { message: string; retry: () => void }) {
  return <div className="error-state" role="alert"><span className="error-mark" aria-hidden="true">!</span><span className="eyebrow">No pudimos cargar la información</span><h2>Probemos de nuevo.</h2><p>{message}</p><button className="button primary" onClick={retry}>Reintentar</button></div>;
}
