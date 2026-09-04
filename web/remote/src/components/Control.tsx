import React, { useEffect } from "react";
import KeyboardArrowLeftIcon from "@mui/icons-material/KeyboardArrowLeft";
import KeyboardArrowRightIcon from "@mui/icons-material/KeyboardArrowRight";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import VolumeUpIcon from "@mui/icons-material/VolumeUp";
import VolumeDownIcon from "@mui/icons-material/VolumeDown";
import VolumeOffIcon from "@mui/icons-material/VolumeOff";
import PersonIcon from "@mui/icons-material/Person";
import RestartAltIcon from "@mui/icons-material/RestartAlt";
import BluetoothIcon from "@mui/icons-material/Bluetooth";
import AddAPhotoIcon from "@mui/icons-material/AddAPhoto";
import KeyboardIcon from "@mui/icons-material/Keyboard";
import MouseIcon from "@mui/icons-material/Mouse";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Grid from "@mui/material/Grid";
import KeyboardDoubleArrowDownIcon from "@mui/icons-material/KeyboardDoubleArrowDown";
import KeyboardDoubleArrowUpIcon from "@mui/icons-material/KeyboardDoubleArrowUp";

import Keyboard from "react-simple-keyboard";
import "react-simple-keyboard/build/css/index.css";
import { Dialog } from "@mui/material";
import useWs from "./WebSocket";
import MousePad from "./MousePad";
import Stack from "@mui/material/Stack";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";

export const keyMap: { [key: string]: number } = {
  ESC: 1,
  "1": 2,
  "2": 3,
  "3": 4,
  "4": 5,
  "5": 6,
  "6": 7,
  "7": 8,
  "8": 9,
  "9": 10,
  "0": 11,
  "-": 12,
  "=": 13,
  "{backspace}": 14,
  "{tab}": 15,
  q: 16,
  w: 17,
  e: 18,
  r: 19,
  t: 20,
  y: 21,
  u: 22,
  i: 23,
  o: 24,
  p: 25,
  "[": 26,
  "]": 27,
  "{enter}": 28,
  a: 30,
  s: 31,
  d: 32,
  f: 33,
  g: 34,
  h: 35,
  j: 36,
  k: 37,
  l: 38,
  ";": 39,
  "'": 40,
  "`": 41,
  "\\": 43,
  z: 44,
  x: 45,
  c: 46,
  v: 47,
  b: 48,
  n: 49,
  m: 50,
  ",": 51,
  ".": 52,
  "/": 53,
  "{space}": 57,
  F1: 59,
  F2: 60,
  F3: 61,
  F4: 62,
  F5: 63,
  F6: 64,
  F7: 65,
  F8: 66,
  F9: 67,
  F10: 68,
  SCRLK: 70,
  PAUSE: 72,
  F11: 87,
  F12: 88,
  HOME: 102,
  "{up}": 103,
  PGUP: 104,
  "{left}": 105,
  "{right}": 106,
  END: 107,
  "{down}": 108,
  PGDN: 109,
  INS: 110,
  DEL: 111,
  PRTSC: 210, // maybe, check this
  // hold shift
  "!": -2,
  "@": -3,
  "#": -4,
  $: -5,
  "%": -6,
  "^": -7,
  "&": -8,
  "*": -9,
  "(": -10,
  ")": -11,
  _: -12,
  "+": -13,
  Q: -16,
  W: -17,
  E: -18,
  R: -19,
  T: -20,
  Y: -21,
  U: -22,
  I: -23,
  O: -24,
  P: -25,
  "{": -26,
  "}": -27,
  A: -30,
  S: -31,
  D: -32,
  F: -33,
  G: -34,
  H: -35,
  J: -36,
  K: -37,
  L: -38,
  ":": -39,
  '"': -40,
  "~": -41,
  "|": -43,
  Z: -44,
  X: -45,
  C: -46,
  V: -47,
  B: -48,
  N: -49,
  M: -50,
  "<": -51,
  ">": -52,
  "?": -53,
};

function isTouchSupported() {
  // @ts-ignore
  const msTouchEnabled = window.navigator.msMaxTouchPoints;
  const generalTouchEnabled = "ontouchstart" in document.createElement("div");
  return !!(msTouchEnabled || generalTouchEnabled);
}

export default function Control() {
  const [keyboardLayout, setKeyboardLayout] = React.useState("default");
  const [keyboardOpen, setKeyboardOpen] = React.useState(false);
  const [numpadOpen, setNumpadOpen] = React.useState(false);
  const [mouseOpen, setMouseOpen] = React.useState(false);
  const [resetOpen, setResetOpen] = React.useState(false);

  const ws = useWs();

  useEffect(() => {
    window.scrollTo({
      top: 0,
    });
  }, []);

  const sendKey = (key: string) => {
    ws.sendMessage("kbd:" + key);
  };

  const sendRawKey = (code: number) => {
    ws.sendMessage("kbdRaw:" + code);
  };

  const hasTouch = isTouchSupported();

  const sendRawKeyDown = (code: number) => {
    if (hasTouch) {
      ws.sendMessage("kbdRawDown:" + code);
    }
  };

  const sendRawKeyUp = (code: number) => {
    if (hasTouch) {
      ws.sendMessage("kbdRawUp:" + code);
    }
  };

  const sendRawKeyNoTouch = (code: number) => {
    if (!hasTouch) {
      ws.sendMessage("kbdRaw:" + code);
    }
  };

  return (
    <Box margin={2}>
      <Grid container spacing={2}>
        <Grid item xs={4}>
          <Button
            variant="outlined"
            size="large"
            sx={{ width: "100%", height: 75 }}
            onClick={() => {
              sendKey("back");
            }}
          >
            <ArrowBackIcon />
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="contained"
            size="large"
            sx={{ width: "100%", height: 75 }}
            onTouchStart={() => {
              sendRawKeyDown(103);
            }}
            onTouchEnd={() => {
              sendRawKeyUp(103);
            }}
            onTouchCancel={() => {
              sendRawKeyUp(103);
            }}
            onClick={() => {
              sendRawKeyNoTouch(103);
            }}
          >
            <KeyboardArrowUpIcon />
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="outlined"
            size="large"
            sx={{ width: "100%", height: 75 }}
            onClick={() => {
              sendKey("osd");
            }}
          >
            OSD
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="contained"
            size="large"
            sx={{ width: "100%", height: 75 }}
            onTouchStart={() => {
              sendRawKeyDown(105);
            }}
            onTouchEnd={() => {
              sendRawKeyUp(105);
            }}
            onTouchCancel={() => {
              sendRawKeyUp(105);
            }}
            onClick={() => {
              sendRawKeyNoTouch(105);
            }}
          >
            <KeyboardArrowLeftIcon />
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="contained"
            size="large"
            sx={{ width: "100%", height: 75 }}
            onClick={() => {
              sendKey("confirm");
            }}
          >
            OK
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="contained"
            size="large"
            sx={{ width: "100%", height: 75 }}
            onTouchStart={() => {
              sendRawKeyDown(106);
            }}
            onTouchEnd={() => {
              sendRawKeyUp(106);
            }}
            onTouchCancel={() => {
              sendRawKeyUp(106);
            }}
            onClick={() => {
              sendRawKeyNoTouch(106);
            }}
          >
            <KeyboardArrowRightIcon />
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="outlined"
            size="large"
            sx={{ width: "100%", height: 75 }}
            onClick={() => {
              sendKey("cancel");
            }}
          >
            Cancel
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="contained"
            size="large"
            sx={{ width: "100%", height: 75 }}
            onTouchStart={() => {
              sendRawKeyDown(108);
            }}
            onTouchEnd={() => {
              sendRawKeyUp(108);
            }}
            onTouchCancel={() => {
              sendRawKeyUp(108);
            }}
            onClick={() => {
              sendRawKeyNoTouch(108);
            }}
          >
            <KeyboardArrowDownIcon />
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Stack direction="row" spacing={0.5}>
            <Button
              variant="outlined"
              size="large"
              sx={{ height: 75, flexGrow: 1, minWidth: 0 }}
              onClick={() => {
                sendRawKey(104);
              }}
            >
              <KeyboardDoubleArrowUpIcon />
            </Button>
            <Button
              variant="outlined"
              size="large"
              sx={{ height: 75, flexGrow: 1, minWidth: 0 }}
              onClick={() => {
                sendRawKey(109);
              }}
            >
              <KeyboardDoubleArrowDownIcon />
            </Button>
          </Stack>
        </Grid>
      </Grid>

      <Grid container spacing={2} sx={{ mt: 3 }}>
        <Grid item xs={4}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendRawKey(2);
            }}
          >
            1
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendRawKey(57);
            }}
          >
            Space
          </Button>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendRawKey(6);
            }}
          >
            5
          </Button>
        </Grid>
      </Grid>

      <Grid container spacing={2} sx={{ mt: 3 }}>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              setKeyboardOpen(true);
            }}
            startIcon={<KeyboardIcon />}
          >
            Keyboard
          </Button>
        </Grid>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              setMouseOpen(true);
            }}
            startIcon={<MouseIcon />}
          >
            Mouse
          </Button>
        </Grid>

        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              setNumpadOpen(true);
            }}
          >
            Keypad
          </Button>
        </Grid>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendKey("computer_osd");
            }}
          >
            Comp. OSD
          </Button>
        </Grid>

        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendKey("volume_up");
            }}
            startIcon={<VolumeUpIcon />}
          >
            Vol. up
          </Button>
        </Grid>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendKey("screenshot");
            }}
            startIcon={<AddAPhotoIcon />}
          >
            Screenshot
          </Button>
        </Grid>

        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendKey("volume_down");
            }}
            startIcon={<VolumeDownIcon />}
          >
            Vol. down
          </Button>
        </Grid>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendKey("raw_screenshot");
            }}
            startIcon={<AddAPhotoIcon />}
          >
            Raw
          </Button>
        </Grid>

        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendKey("volume_mute");
            }}
            startIcon={<VolumeOffIcon />}
          >
            Mute
          </Button>
        </Grid>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendKey("user");
            }}
            startIcon={<PersonIcon />}
          >
            User
          </Button>
        </Grid>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              sendKey("pair_bluetooth");
            }}
            startIcon={<BluetoothIcon />}
          >
            Pair BT
          </Button>
        </Grid>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            sx={{ width: "100%" }}
            onClick={() => {
              setResetOpen(true);
            }}
            startIcon={<RestartAltIcon />}
          >
            Reset
          </Button>
        </Grid>
      </Grid>

      <MousePad
        open={mouseOpen}
        onClose={() => setMouseOpen(false)}
        sendMessage={ws.sendMessage}
      />

      <Dialog open={resetOpen} onClose={() => setResetOpen(false)}>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to reset the system?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setResetOpen(false)}>Cancel</Button>
          <Button
            onClick={() => {
              sendKey("reset");
              setResetOpen(false);
            }}
            autoFocus
            variant="contained"
            color="error"
          >
            Reset
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        onClose={() => setKeyboardOpen(false)}
        open={keyboardOpen}
        fullWidth
        PaperProps={{
          sx: {
            position: "fixed",
            bottom: 10,
            m: 0,
            maxWidth: "95%",
            width: "95%",
          },
        }}
      >
        <div style={{ color: "black" }}>
          <Keyboard
            layoutName={keyboardLayout}
            layout={{
              default: [
                "q w e r t y u i o p",
                "a s d f g h j k l",
                "{symbols} z x c v b n m {backspace}",
                "{numbers} , {space} . {shiftup} {enter}",
              ],
              shift: [
                "Q W E R T Y U I O P",
                "A S D F G H J K L",
                "{symbols} Z X C V B N M {backspace}",
                "{numbers} , {space} . {shiftdown} {enter}",
              ],
              numbers: [
                "1 2 3 4 5 6 7 8 9 0",
                "@ # $ _ & - + ( ) /",
                "{symbols} * \" ' : ; ! ? {backspace}",
                "{abc} , {space} . {enter}",
              ],
              symbols: [
                "ESC F1 F2 F3 F4 F5 F6",
                "{tab} F7 F8 F9 F10 F11 F12",
                "PRTSC SCRLK PAUSE INS",
                "DEL PGUP PGDN HOME",
                "{left} {up} {down} {right} END",
                "^ < > = { } ~ | \\",
                "{abc} % [ ] ` {backspace}",
                "{numbers} , {space} . {enter}",
              ],
            }}
            display={{
              "{numbers}": "123",
              "{enter}": "↵",
              "{tab}": "⇥",
              "{backspace}": "⌫",
              "{shiftup}": "⇧",
              "{shiftdown}": "⬆",
              "{abc}": "ABC",
              "{space}": "_________",
              "{symbols}": "FUNC",
              "{left}": "←",
              "{right}": "→",
              "{up}": "↑",
              "{down}": "↓",
            }}
            onKeyPress={(input) => {
              if (input === "{shiftup}") {
                setKeyboardLayout("shift");
                return;
              } else if (input === "{shiftdown}") {
                setKeyboardLayout("default");
                return;
              } else if (input === "{numbers}") {
                setKeyboardLayout("numbers");
                return;
              } else if (input === "{symbols}") {
                setKeyboardLayout("symbols");
                return;
              } else if (input === "{abc}") {
                setKeyboardLayout("default");
                return;
              }

              if (input in keyMap) {
                sendRawKey(keyMap[input]);
              }
            }}
          />
        </div>
      </Dialog>

      <Dialog
        onClose={() => setNumpadOpen(false)}
        open={numpadOpen}
        fullWidth
        PaperProps={{
          sx: {
            position: "fixed",
            bottom: 10,
            m: 0,
            // maxWidth: "95%",
            // width: "95%",
          },
        }}
      >
        <div style={{ color: "black" }}>
          <Keyboard
            layout={{
              default: ["1 2 3", "4 5 6", "7 8 9", "* 0 #"],
            }}
            onKeyPress={(input) => {
              if (input in keyMap) {
                sendRawKey(keyMap[input]);
              }
            }}
          />
        </div>
      </Dialog>
    </Box>
  );
}
