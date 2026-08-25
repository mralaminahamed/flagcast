import { Link, useRouteError } from "react-router-dom";

export function RouteError() {
  const err = useRouteError() as Error | undefined;
  return (
    <div className="mx-auto max-w-lg px-6 py-20 text-center">
      <h1 className="font-mono text-xl font-semibold">Something went wrong</h1>
      <p className="mt-2 break-words text-sm text-muted">{err?.message ?? "Unexpected error."}</p>
      <Link to="/" className="mt-5 inline-block rounded-md border border-line px-4 py-2 text-sm hover:border-brand">
        Back to flags
      </Link>
    </div>
  );
}

export function NotFound() {
  return (
    <div className="mx-auto max-w-lg px-6 py-20 text-center">
      <h1 className="font-mono text-xl font-semibold">Page not found</h1>
      <Link to="/" className="mt-5 inline-block rounded-md border border-line px-4 py-2 text-sm hover:border-brand">
        Back to flags
      </Link>
    </div>
  );
}
