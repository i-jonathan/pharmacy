import { inject } from "vue";

export const PermissionsKey = Symbol("permissions");
export const UserKey = Symbol("user");

export function usePermissions() {
  const perms = inject(PermissionsKey, {});
  const hasPermission = (key) => !!perms.value?.[key] ?? !!perms[key];
  return { permissions: perms, hasPermission };
}

export function useUser() {
  return inject(UserKey, { id: 0 });
}
