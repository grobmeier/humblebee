type HumbleBeeLogoProps = {
  className?: string;
};

export function HumbleBeeLogo({ className }: HumbleBeeLogoProps) {
  return (
    <svg
      aria-hidden="true"
      className={className}
      focusable="false"
      viewBox="0 0 48 48"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M22.3 13.2c3.7-5.4 10.9-6.7 15.2-2.6-2.7 5.2-9.9 7.1-15.2 2.6Z"
        fill="currentColor"
        opacity="0.86"
      />
      <path
        d="M10.6 22.4c0-3.2 2.6-5.8 5.8-5.8h2.1v11.6h-2.1c-3.2 0-5.8-2.6-5.8-5.8Z"
        fill="currentColor"
      />
      <path
        d="M17.6 15.4h12.7c6.6 0 12 4 12 9.8s-5.4 9.8-12 9.8H17.6V15.4Z"
        fill="currentColor"
      />
      <path
        d="M43.2 25.2 38 21.8v6.8l5.2-3.4Z"
        fill="currentColor"
      />
      <path
        d="M24.6 16v17.9M32.2 16.8v16.1"
        fill="none"
        stroke="#0a85c5"
        strokeLinecap="round"
        strokeWidth="3"
      />
      <path
        d="M11.3 18.6 7.8 15M12.1 16.8l-1.8-4"
        fill="none"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="2.2"
      />
      <circle cx="14.7" cy="20.7" fill="#0a85c5" r="1.2" />
    </svg>
  );
}
