import { ThemeOptions } from "@mui/material";

interface ThemeDefinition {
  id: string;
  displayName: string;
  options: ThemeOptions;
}

export const themes: { [id: string]: ThemeDefinition } = {
  tron: {
    id: "tron",
    displayName: "Tron",
    options: {
      palette: {
        mode: "dark",
        primary: {
          main: "#13e8ec",
        },
        secondary: {
          main: "#13e8ec",
        },
        text: {
          primary: "#13e8ec",
          secondary: "#0a6b6d",
        },
      },
    },
  },
  mister: {
    id: "mister",
    displayName: "MiSTer",
    options: {
      palette: {
        mode: "dark",
        background: {
          default: "#282a36",
        },
        primary: {
          main: "#9c5cf7",
          contrastText: "#fff",
        },
        secondary: {
          main: "#514689",
        },
        text: {
          primary: "#fff",
        },
      },
    },
  },
  misterLight: {
    id: "misterLight",
    displayName: "MiSTer Light",
    options: {
      palette: {
        mode: "light",
        background: {
          default: "#ffffff",
        },
        primary: {
          main: "#9c5cf7",
          contrastText: "#000000",
        },
        secondary: {
          main: "#c1bcdc",
        },
        text: {
          primary: "#000000",
        },
      },
    },
  },
  misterConsoles: {
    id: "misterConsoles",
    displayName: "MiSTer Consoles",
    options: {
      palette: {
        mode: "dark",
        primary: {
          main: "#257cc7",
        },
        secondary: {
          main: "#d31e2b",
        },
        warning: {
          main: "#efa400",
        },
        success: {
          main: "#2bc32b",
        },
      },
    },
  },
};

export function getTheme(id: string): ThemeDefinition {
  if (themes[id]) {
    return themes[id];
  } else {
    return themes.mister;
  }
}
