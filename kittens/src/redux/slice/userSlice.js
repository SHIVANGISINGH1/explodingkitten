import { createSlice } from "@reduxjs/toolkit";
import io from "socket.io-client";
import { getItem } from "../../utils/storage";
import { request } from "../../api/request";
import { fetchUsers } from "./leaderboardSlice";
const ENDPOINT = "localhost:4000";
var socket;

const userSlice = createSlice({
  name: "user",
  initialState: {
    defuseCards: [],
    openedCard: "",
    deck: [],
    username: "",
    matchesWon: 0,
  },
  reducers: {
    get_user(state, action) {
      console.log("state = ", state);
      console.log("action = ", action);
      if (action.payload.matchesWon == null) {
        console.log("love", state.matchesWon);
      }
      if (action.payload?.deck) {
        return action.payload;
      } else {
        return {
          ...state,
          username: action.payload?.username,
          matchesWon:
            action.payload?.matchesWon == null
              ? state.matchesWon
              : action.payload?.matchesWon,
        };
      }
    },
    save_won_game(state, action) {
      return { ...state, user: action.payload };
    },
  },
});

export default userSlice.reducer;
export const { get_user, save_won_game } = userSlice.actions;

export function getUserDetail() {
  return async function getUserDataThunk(dispatch, getstate) {
    const user = getItem("user");
    dispatch(get_user(user));
  };
}

export function editUserDetail(userData) {
  console.log("userdate for dispatch", userData);
  return async function editUserDataThunk(dispatch, getstate) {
    localStorage.setItem("user", JSON.stringify(userData));
    console.log("listt", userData);
    dispatch(get_user(userData));
  };
}

export function updateUserPoints(userData) {
  return async function updateUserPointsThunk(dispatch, getstate) {
    let reqbody = JSON.stringify({
      username: userData.username,
      matchesWon: userData.matchesWon.toString(),
    });
    let response = await request("PUT", "/updateUser", reqbody);

    if (response.status == 200) {
      socket = io(ENDPOINT);
      dispatch(fetchUsers());
      socket.emit("updateLeaderBoard");
    }
  };
}
