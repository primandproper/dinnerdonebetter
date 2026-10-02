// See https://svelte.dev/docs/kit/types#app.d.ts
import type { Session } from '@primandproper/platform-client';

declare global {
  namespace App {
    interface Locals {
      session: Session;
    }
  }
}

export {};
