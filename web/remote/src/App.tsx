import React from "react";

import { createTheme, ThemeProvider } from "@mui/material/styles";
import { getTheme } from "./lib/themes";

import Layout from "./components/Layout";
import { useUIStateStore } from "./lib/store";

function App() {
  const activeThemeState = useUIStateStore((state) => state.activeTheme);
  const activeTheme = getTheme(activeThemeState);
  const fontSize = useUIStateStore((state) => state.fontSize);

  const theme = createTheme({
    ...activeTheme.options,
    typography: {
      ...activeTheme.options.typography,
      fontSize: fontSize,
    },
  });

  return (
    <ThemeProvider theme={theme}>
      <Layout />
    </ThemeProvider>
  );
}

export default App;
