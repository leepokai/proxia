import * as React from "react";
import { cn } from "./utils";

export type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "default" | "outline" | "ghost";
};

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", ...props }, ref) => {
    const base =
      "inline-flex items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none focus:ring-offset-neutral-950";
    const styles: Record<NonNullable<ButtonProps["variant"]>, string> = {
      default: "bg-indigo-500 text-white hover:bg-indigo-400 focus:ring-indigo-400",
      outline:
        "border border-neutral-700 text-neutral-100 hover:bg-neutral-800 focus:ring-neutral-700",
      ghost: "text-neutral-200 hover:bg-neutral-800 focus:ring-neutral-700",
    };
    return (
      <button
        ref={ref}
        className={cn(base, styles[variant], className)}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";




