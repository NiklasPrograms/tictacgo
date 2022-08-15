import { BackendSquareChacter, BackendWinningCharacter } from "../adapter/Adapter";
import { Position } from "../utility/Position";

export const BASE_URL = "localhost:8080";
export const GAME_WS_URL = `ws://${BASE_URL}/gamews`;
export const CHAT_WS_URL = `ws://${BASE_URL}/chatws`;

export enum GameCommand {
    RESULT = "result",
    GAME_OVER = "game over",
    BOARD = "board",
};

export type JSONResult = {
    WinningCombination: Position[],
    WinningCharacter: BackendWinningCharacter,
    HasWinner: boolean
}

export type GameBody = JSONResult | BackendSquareChacter[] | BackendSquareChacter | boolean

// Not used yet, but should be used to validate the response from the game.
export type GameResponse = {
    command: GameCommand,
    body: GameBody
};
