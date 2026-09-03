import Grid from "@mui/material/Grid";
import ArrowUpwardIcon from "@mui/icons-material/ArrowUpward";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";
import ArrowDownwardIcon from "@mui/icons-material/ArrowDownward";
import React, { useEffect, useState } from "react";
import Button from "@mui/material/Button";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import { useUIStateStore } from "../lib/store";
import TextField from "@mui/material/TextField";
import { keyMap } from "./Control";
import Divider from "@mui/material/Divider";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import KeyboardArrowRightIcon from "@mui/icons-material/KeyboardArrowRight";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowLeftIcon from "@mui/icons-material/KeyboardArrowLeft";
import useWs from "./WebSocket";
import moment from "moment";

function CircleButton(props: {
  children: React.ReactNode;
  clickHandler: () => void;
  id: string;
}) {
  return (
    <Button
      id={props.id}
      variant="contained"
      sx={{ borderRadius: "100%", minWidth: 0, width: "80px", height: "80px" }}
      onClick={props.clickHandler}
    >
      {props.children}
    </Button>
  );
}

function ActionButtons(props: {
  sendKey: (key: number) => void;
  logKey: (name: string, code: number) => void;
  keys: number[];
}) {
  return (
    <Grid container>
      <Grid item xs={4}></Grid>
      <Grid item xs={4}>
        <CircleButton
          id="button-x"
          clickHandler={() => {
            props.sendKey(props.keys[11]);
            props.logKey("X", props.keys[11]);
          }}
        >
          <Typography fontSize="x-large">X</Typography>
        </CircleButton>
      </Grid>
      <Grid item xs={4}></Grid>
      <Grid item xs={4}>
        <CircleButton
          id="button-y"
          clickHandler={() => {
            props.sendKey(props.keys[12]);
            props.logKey("Y", props.keys[12]);
          }}
        >
          <Typography fontSize="x-large">Y</Typography>
        </CircleButton>
      </Grid>
      <Grid item xs={4}></Grid>
      <Grid item xs={4}>
        <CircleButton
          id="button-a"
          clickHandler={() => {
            props.sendKey(props.keys[14]);
            props.logKey("A", props.keys[14]);
          }}
        >
          <Typography fontSize="x-large">A</Typography>
        </CircleButton>
      </Grid>
      <Grid item xs={4}></Grid>
      <Grid item xs={4}>
        <CircleButton
          id="button-b"
          clickHandler={() => {
            props.sendKey(props.keys[13]);
            props.logKey("B", props.keys[13]);
          }}
        >
          <Typography fontSize="x-large">B</Typography>
        </CircleButton>
      </Grid>
      <Grid item xs={4}></Grid>
    </Grid>
  );
}

function DirectionButtons(props: {
  sendKey: (key: number) => void;
  logKey: (name: string, code: number) => void;
  keys: number[];
}) {
  return (
    <Grid container>
      <Grid item xs={4}></Grid>
      <Grid item xs={4}>
        <CircleButton
          id="button-dpad-up"
          clickHandler={() => {
            props.sendKey(props.keys[4]);
            props.logKey("Dpad up", props.keys[4]);
          }}
        >
          <ArrowUpwardIcon fontSize="large" />
        </CircleButton>
      </Grid>
      <Grid item xs={4}></Grid>
      <Grid item xs={4}>
        <CircleButton
          id="button-dpad-left"
          clickHandler={() => {
            props.sendKey(props.keys[6]);
            props.logKey("Dpad left", props.keys[6]);
          }}
        >
          <ArrowBackIcon fontSize="large" />
        </CircleButton>
      </Grid>
      <Grid item xs={4}></Grid>
      <Grid item xs={4}>
        <CircleButton
          id="button-dpad-right"
          clickHandler={() => {
            props.sendKey(props.keys[7]);
            props.logKey("Dpad right", props.keys[7]);
          }}
        >
          <ArrowForwardIcon fontSize="large" />
        </CircleButton>
      </Grid>
      <Grid item xs={4}></Grid>
      <Grid item xs={4}>
        <CircleButton
          id="button-dpad-down"
          clickHandler={() => {
            props.sendKey(props.keys[5]);
            props.logKey("Dpad down", props.keys[5]);
          }}
        >
          <ArrowDownwardIcon fontSize="large" />
        </CircleButton>
      </Grid>
      <Grid item xs={4}></Grid>
    </Grid>
  );
}

function MiddleButtons(props: {
  sendKey: (key: number) => void;
  logKey: (name: string, code: number) => void;
  keys: number[];
}) {
  return (
    <Grid container>
      <Grid item xs={12}>
        <CircleButton
          id="button-pad-osd"
          clickHandler={() => {
            props.sendKey(props.keys[8]);
            props.logKey("Pad OSD", props.keys[8]);
          }}
        >
          OSD
        </CircleButton>
      </Grid>
      <Grid item xs={6}>
        <Button
          id="button-select"
          variant="contained"
          sx={{ width: "80%", mt: 8, borderRadius: 10 }}
          onClick={() => {
            props.sendKey(props.keys[9]);
            props.logKey("Select", props.keys[9]);
          }}
        >
          Select
        </Button>
      </Grid>
      <Grid item xs={6}>
        <Button
          id="button-start"
          variant="contained"
          sx={{ width: "80%", mt: 8, borderRadius: 10 }}
          onClick={() => {
            props.sendKey(props.keys[10]);
            props.logKey("Start", props.keys[10]);
          }}
        >
          Start
        </Button>
      </Grid>
    </Grid>
  );
}

export function ControlAuto() {
  const uiState = useUIStateStore();
  const ws = useWs();

  const [log, setLog] = useState<string>("");

  const logKey = (name: string, code: number) => {
    setLog(
      moment().format("YYYY-MM-DD HH:mm:ss.S") +
        " - " +
        name +
        " (" +
        code +
        ")\n" +
        log
    );
  };

  const logGeneric = (msg: string) => {
    setLog(moment().format("YYYY-MM-DD HH:mm:ss.S") + " - " + msg + "\n" + log);
  };

  const editKey = (index: number, value: string) => {
    const newKeys = [...uiState.autoControlKeys];
    newKeys[index] = parseInt(value, 10);
    uiState.setAutoControlKeys(newKeys);
  };

  const sendKey = (key: string) => {
    ws.sendMessage("kbd:" + key);
  };

  const sendRawKey = (code: number) => {
    ws.sendMessage("kbdRaw:" + code);
  };

  useEffect(() => {
    if (!ws.lastMessage || !ws.lastMessage.data) return;
    const data = ws.lastMessage.data;
    if (data.startsWith("gameRunning:")) {
      let gameRunning = data.split(":")[1];
      if (gameRunning.length === 0) {
        gameRunning = "<NONE>";
      }
      logGeneric("Game running: " + gameRunning);
    } else if (data.startsWith("coreRunning:")) {
      let coreRunning = data.split(":")[1];
      if (coreRunning.length === 0) {
        coreRunning = "<NONE>";
      }
      logGeneric("Core running: " + coreRunning);
    }
  }, [ws.lastMessage]);

  return (
    <Box>
      <Box
        sx={{ width: "100%", display: "flex", justifyContent: "space-around" }}
      >
        <Grid container sx={{ textAlign: "center", p: 2, maxWidth: "800px" }}>
          <Grid item xs={2}>
            <Button
              id="button-l2"
              variant="contained"
              sx={{ width: "90%", mb: 2 }}
              onClick={() => {
                sendRawKey(uiState.autoControlKeys[0]);
                logKey("L2", uiState.autoControlKeys[0]);
              }}
            >
              L2
            </Button>
          </Grid>
          <Grid item xs={2}>
            <Button
              id="button-l1"
              variant="contained"
              sx={{ width: "100%", mb: 2 }}
              onClick={() => {
                sendRawKey(uiState.autoControlKeys[1]);
                logKey("L1", uiState.autoControlKeys[1]);
              }}
            >
              L1
            </Button>
          </Grid>
          <Grid item xs={4}></Grid>
          <Grid item xs={2}>
            <Button
              id="button-r1"
              variant="contained"
              sx={{ width: "100%", mb: 2 }}
              onClick={() => {
                sendRawKey(uiState.autoControlKeys[2]);
                logKey("R1", uiState.autoControlKeys[2]);
              }}
            >
              R1
            </Button>
          </Grid>
          <Grid item xs={2}>
            <Button
              id="button-r2"
              variant="contained"
              sx={{ width: "90%", mb: 2 }}
              onClick={() => {
                sendRawKey(uiState.autoControlKeys[3]);
                logKey("R2", uiState.autoControlKeys[3]);
              }}
            >
              R2
            </Button>
          </Grid>
          <Grid item xs={4}>
            <DirectionButtons
              keys={uiState.autoControlKeys}
              logKey={logKey}
              sendKey={sendRawKey}
            />
          </Grid>
          <Grid item xs={4} sx={{ pl: 2, pr: 2 }}>
            <MiddleButtons
              keys={uiState.autoControlKeys}
              logKey={logKey}
              sendKey={sendRawKey}
            />
          </Grid>
          <Grid item xs={4}>
            <ActionButtons
              keys={uiState.autoControlKeys}
              logKey={logKey}
              sendKey={sendRawKey}
            />
          </Grid>
          <Grid item xs={12}>
            <Stack
              direction="row"
              spacing={1}
              sx={{ mt: 5, mb: 1, justifyContent: "space-between" }}
            >
              <Button
                id="button-1"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[15]);
                  logKey("1", uiState.autoControlKeys[15]);
                }}
              >
                1
              </Button>
              <Button
                id="button-2"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[16]);
                  logKey("2", uiState.autoControlKeys[16]);
                }}
              >
                2
              </Button>
              <Button
                id="button-3"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[17]);
                  logKey("3", uiState.autoControlKeys[17]);
                }}
              >
                3
              </Button>
              <Button
                id="button-4"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[18]);
                  logKey("4", uiState.autoControlKeys[18]);
                }}
              >
                4
              </Button>
              <Button
                id="button-5"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[19]);
                  logKey("5", uiState.autoControlKeys[19]);
                }}
              >
                5
              </Button>
            </Stack>
          </Grid>
          <Grid item xs={12}>
            <Stack
              direction="row"
              spacing={1}
              sx={{ mb: 1, justifyContent: "space-between" }}
            >
              <Button
                id="button-6"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[20]);
                  logKey("6", uiState.autoControlKeys[20]);
                }}
              >
                6
              </Button>
              <Button
                id="button-7"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[21]);
                  logKey("7", uiState.autoControlKeys[21]);
                }}
              >
                7
              </Button>
              <Button
                id="button-8"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[22]);
                  logKey("8", uiState.autoControlKeys[22]);
                }}
              >
                8
              </Button>
              <Button
                id="button-9"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[23]);
                  logKey("9", uiState.autoControlKeys[23]);
                }}
              >
                9
              </Button>
              <Button
                id="button-10"
                variant="outlined"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendRawKey(uiState.autoControlKeys[24]);
                  logKey("10", uiState.autoControlKeys[24]);
                }}
              >
                10
              </Button>
            </Stack>
          </Grid>
          <Grid item xs={12}>
            <Stack
              direction="row"
              spacing={1}
              sx={{ justifyContent: "space-between" }}
            >
              <Button
                id="button-mister-up"
                variant="contained"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendKey("up");
                  logKey("MiSTer up", 0);
                }}
              >
                <KeyboardArrowUpIcon />
              </Button>
              <Button
                id="button-mister-down"
                variant="contained"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendKey("down");
                  logKey("MiSTer down", 0);
                }}
              >
                <KeyboardArrowDownIcon />
              </Button>
              <Button
                id="button-mister-left"
                variant="contained"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendKey("left");
                  logKey("MiSTer left", 0);
                }}
              >
                <KeyboardArrowLeftIcon />
              </Button>
              <Button
                id="button-mister-right"
                variant="contained"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendKey("right");
                  logKey("MiSTer right", 0);
                }}
              >
                <KeyboardArrowRightIcon />
              </Button>
              <Button
                id="button-mister-ok"
                variant="contained"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendKey("confirm");
                  logKey("MiSTer ok", 0);
                }}
              >
                OK
              </Button>
              <Button
                id="button-mister-back"
                variant="contained"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendKey("back");
                  logKey("MiSTer back", 0);
                }}
              >
                Back
              </Button>
              <Button
                id="button-mister-cancel"
                variant="contained"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendKey("cancel");
                  logKey("MiSTer cancel", 0);
                }}
              >
                Cancel
              </Button>
              <Button
                id="button-mister-osd"
                variant="contained"
                sx={{ width: "100%" }}
                onClick={() => {
                  sendKey("osd");
                  logKey("MiSTer OSD", 0);
                }}
              >
                OSD
              </Button>
            </Stack>
          </Grid>
        </Grid>
      </Box>
      <Divider sx={{ m: 1 }} />
      <Grid container spacing={1} p={1}>
        <Grid item xs={4}>
          <TextField
            variant="outlined"
            multiline
            sx={{ width: "100%" }}
            disabled
            minRows={20}
            maxRows={20}
            value={log}
          />
        </Grid>
        <Grid item xs={8}>
          <Stack spacing={2}>
            <Grid container spacing={1}>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="L2"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[0]}
                  onChange={(e) => {
                    editKey(0, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="L1"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[1]}
                  onChange={(e) => {
                    editKey(1, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="R1"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[2]}
                  onChange={(e) => {
                    editKey(2, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="R2"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[3]}
                  onChange={(e) => {
                    editKey(3, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="Up"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[4]}
                  onChange={(e) => {
                    editKey(4, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="Down"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[5]}
                  onChange={(e) => {
                    editKey(5, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="Left"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[6]}
                  onChange={(e) => {
                    editKey(6, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="Right"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[7]}
                  onChange={(e) => {
                    editKey(7, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="OSD"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[8]}
                  onChange={(e) => {
                    editKey(8, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="Select"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[9]}
                  onChange={(e) => {
                    editKey(9, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="Start"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[10]}
                  onChange={(e) => {
                    editKey(10, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="X"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[11]}
                  onChange={(e) => {
                    editKey(11, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="Y"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[12]}
                  onChange={(e) => {
                    editKey(12, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="B"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[13]}
                  onChange={(e) => {
                    editKey(13, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="A"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[14]}
                  onChange={(e) => {
                    editKey(14, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="1"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[15]}
                  onChange={(e) => {
                    editKey(15, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="2"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[16]}
                  onChange={(e) => {
                    editKey(16, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="3"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[17]}
                  onChange={(e) => {
                    editKey(17, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="4"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[18]}
                  onChange={(e) => {
                    editKey(18, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="5"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[19]}
                  onChange={(e) => {
                    editKey(19, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="6"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[20]}
                  onChange={(e) => {
                    editKey(20, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="7"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[21]}
                  onChange={(e) => {
                    editKey(21, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="8"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[22]}
                  onChange={(e) => {
                    editKey(22, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="9"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[23]}
                  onChange={(e) => {
                    editKey(23, e.target.value);
                  }}
                />
              </Grid>
              <Grid item xs={2}>
                <TextField
                  id="outlined-basic"
                  label="10"
                  variant="outlined"
                  size="small"
                  value={uiState.autoControlKeys[24]}
                  onChange={(e) => {
                    editKey(24, e.target.value);
                  }}
                />
              </Grid>
            </Grid>
            <Box>
              <>{"<KEY> = <UINPUT CODE>"}</>
              {Object.keys(keyMap).map((k: string) => (
                <span key={k}>
                  , {k} = {keyMap[k]}{" "}
                </span>
              ))}
            </Box>
          </Stack>
        </Grid>
      </Grid>
    </Box>
  );
}
