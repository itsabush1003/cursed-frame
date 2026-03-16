import { createContext } from "react";

export type UserType = "admin" | "guest";
export interface UserStatus {
  type: UserType;
  token: string;
  teamId: number;
  color: string;
}

const createUserStatusOrigin = () => {
  const path = window.location.pathname;

  return {
    type: (path.includes("/admin/") ? "admin" : "guest") as UserType,
    token: "",
    teamId: 0,
    color: "transparent",
  };
};
const userStatusOrigin = createUserStatusOrigin();
const setUserStatus = (newUserStatus: Partial<UserStatus>) => {
  Object.assign(userStatusOrigin, newUserStatus);
};

export const UserStatusContext = createContext<{
  userStatus: UserStatus;
  setUserStatus: (newUserStatus: Partial<UserStatus>) => void;
}>({ userStatus: userStatusOrigin, setUserStatus: setUserStatus });
