import { useEffect, useMemo, useState } from "react"
import { useGameContext } from "../../../../context/GameContext"
import { Position } from "../../../../utility/Position"
import "./Square.css"

export type Border = {
    top?: boolean,
    bottom?: boolean,
    left?: boolean,
    right?: boolean
}

enum BorderClass {
    NO_TOP = "no-top-border",
    NO_BOTTOM = "no-bottom-border",
    NO_LEFT = "no-left-border",
    NO_RIGHT = "no-right-border",
    TOP_RIGHT_BORDER_RADIUS = "top-right-border-radius",
    BOTTOM_RIGHT_BORDER_RADIUS = "bottom-right-border-radius",
}

export enum SquareCharacter {
    X = "X",
    O = "O",
}

type EmptyString = "";

export type SquareType = {
    position: Position,
    character: SquareCharacter
}

type SquareProps = {
    position: Position,
    border: Border,
}

// TODO overvej React.memo
// Problemet er, at React.memo ingen effekt har pga useGameContext, men måske man alligevel kan lave noget ala arePropsEqual og så tjek på
// latestSquare.position og winningCombination.
export const Square: React.FC<SquareProps> = (props) => {
    const [character, setCharacter] = useState<SquareCharacter | EmptyString>("");
    const { latestSquare, winningCombination, chooseSquare } = useGameContext();
    
    useEffect(() => {
        const squareHasBeenSelected = latestSquare?.position === props.position;
        if (squareHasBeenSelected) {
            setCharacter(latestSquare.character);  
        }
    }, [latestSquare])

    const borderClasses = useMemo(() => {
        const border = props.border;
        const borderClasses: BorderClass[] = []
        
        if (border.top) {
            borderClasses.push(BorderClass.NO_TOP);
        }
        if (border.bottom) {
            borderClasses.push(BorderClass.NO_BOTTOM);
        }
        if (border.left) {
            borderClasses.push(BorderClass.NO_LEFT);
        }
        if (border.right) {
            borderClasses.push(BorderClass.NO_RIGHT);
        }

        if (border.top && border.right) {
            borderClasses.push(BorderClass.TOP_RIGHT_BORDER_RADIUS);
        }

        if (border.bottom && border.right) {
            borderClasses.push(BorderClass.BOTTOM_RIGHT_BORDER_RADIUS);
        }

        return borderClasses.join(" ");
    }, []);

    const winnerClass = useMemo(() => {
        const isSquareInWinningCombination = winningCombination?.includes(props.position);
        if (isSquareInWinningCombination) {
            return "winner";
        }

        return "";
    }, [winningCombination]);

    return <div onClick={() => chooseSquare(props.position)} className={`square ${character} ${borderClasses} ${winnerClass}`}>
        {character}
    </div>
}