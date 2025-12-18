import * as React from "react";
import { cn } from "./utils";

export type TextareaProps = React.TextareaHTMLAttributes<HTMLTextAreaElement>;

export const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => {
    return (
      <textarea
        ref={ref}
        className={cn(
          "w-full rounded-md border border-neutral-700 bg-neutral-900 px-3 py-2 text-sm text-neutral-100 shadow-sm outline-none ring-0 transition focus:border-neutral-500 focus:ring-2 focus:ring-neutral-700 disabled:cursor-not-allowed disabled:bg-neutral-800 placeholder:text-neutral-500",
          className
        )}
        {...props}
      />
    );
  }
);
Textarea.displayName = "Textarea";




