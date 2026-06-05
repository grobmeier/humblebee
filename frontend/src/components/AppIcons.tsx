type IconProps = {
  className?: string;
};

export function ImportIcon({ className }: IconProps) {
  return (
    <svg className={className} aria-hidden="true" focusable="false" viewBox="0 0 24 24">
      <path d="M12 3v11" />
      <path d="m8 10 4 4 4-4" />
      <path d="M5 15v4h14v-4" />
    </svg>
  );
}

export function DatabaseSwitchIcon({ className }: IconProps) {
  return (
    <svg className={className} aria-hidden="true" focusable="false" viewBox="0 0 24 24">
      <ellipse cx="8" cy="5" rx="5" ry="2.5" />
      <path d="M3 5v7c0 1.4 2.2 2.5 5 2.5 1.2 0 2.3-.2 3.2-.6" />
      <path d="M3 9c0 1.4 2.2 2.5 5 2.5 1.2 0 2.3-.2 3.2-.6" />
      <path d="M15 12h5" />
      <path d="m18 9 3 3-3 3" />
      <path d="M20 17h-5" />
      <path d="m17 20-3-3 3-3" />
    </svg>
  );
}
