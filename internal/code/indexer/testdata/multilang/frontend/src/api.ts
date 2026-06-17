import { fetchUser } from "./user";

export function getApiBase(): string {
  return "/api";
}

export class ApiClient {
  getUser(id: string): string {
    return fetchUser(id);
  }
}

export interface UserStore {
  save(name: string): void;
}
