import React, { Suspense, useContext, useEffect, useState } from 'react';

import IconButton from '@mui/material/IconButton';
import Button from '@mui/material/Button';
import Tooltip from '@mui/material/Tooltip';
import Dialog from '@mui/material/Dialog';
import Box from '@mui/material/Box';
import List from '@mui/material/List';
import ListSubheader from '@mui/material/ListSubheader';
import ListItem from '@mui/material/ListItem';
import Grid from '@mui/material/Grid';

import ChatBubbleIcon from '@mui/icons-material/ChatBubble';
import VideocamIcon from '@mui/icons-material/Videocam';
import CallIcon from '@mui/icons-material/Call';
import FileCopyIcon from '@mui/icons-material/FileCopy';
import InsertPageBreak from '@mui/icons-material/InsertPageBreak';

import { ExchangeActions, useStyles, IFile, OrderedFiles, targets } from 'awayto/hooks';

import ExchangeContext, { ExchangeContextType } from './ExchangeContext';
import WSTextContext, { WSTextContextType } from '../web_socket/WSTextContext';
import WSCallContext, { WSCallContextType } from '../web_socket/WSCallContext';
import FileSelectionModal from '../files/FileSelectionModal';
import Whiteboard from './Whiteboard';

export function Exchange(_: IComponent): React.JSX.Element {
  const classes = useStyles();

  const [dialog, setDialog] = useState('');
  const [chatOpen, setChatOpen] = useState(true);
  const [fileGroups, setFileGroups] = useState<OrderedFiles[]>([])
  const [sharedFile, setSharedFile] = useState<IFile | undefined>();

  const {
    exchangeId,
    getBookingFiles: {
      data: bookingFilesRequest
    }
  } = useContext(ExchangeContext) as ExchangeContextType;

  const {
    chatLog,
    messagesEnd,
    submitMessageForm,
  } = useContext(WSTextContext) as WSTextContextType;

  const {
    audioOnly,
    connected,
    canStartStop,
    localStreamElement,
    senderStreamsElements,
    setLocalStreamAndBroadcast,
    leaveCall
  } = useContext(WSCallContext) as WSCallContextType;

  useEffect(() => {
    setFileGroups(() => [
      { name: 'Exchange', order: 1, files: bookingFilesRequest?.files || [] }
    ]);
  }, [bookingFilesRequest?.files]);

  return <>

    <Dialog fullScreen fullWidth open={dialog === 'file_selection'}>
      <Suspense>
        <FileSelectionModal fileGroups={fileGroups} closeModal={(selectedFile?: IFile) => {
          if (selectedFile) {
            setSharedFile(selectedFile);
          }
          setDialog('');
        }} />
      </Suspense>
    </Dialog>

    <Grid p={1} sx={{ flex: '1 0 25%', display: chatOpen ? 'flex' : 'none', flexDirection: 'column', maxWidth: '390px' }}>
      <Grid sx={{ flex: 1, overflow: 'auto' }}>
        {chatLog}
        {messagesEnd}
      </Grid>

      <Grid pt={1}>
        {submitMessageForm}
      </Grid>
    </Grid>

    <Grid sx={{ flex: 1, display: 'flex', flexDirection: 'column' }}>

      <Grid sx={{ maxHeight: '150px', backgroundColor: '#333' }}>
        {localStreamElement && localStreamElement}
        {senderStreamsElements && senderStreamsElements}
      </Grid>

      <Grid sx={{ height: localStreamElement || senderStreamsElements ? 'calc(100% - 150px)' : '100%', display: 'flex' }}>
        <Whiteboard
          topicId={`exchange/${ExchangeActions.EXCHANGE_WHITEBOARD}:${exchangeId}`}
          sharedFile={sharedFile}
          openFileSelect={() => {
            setDialog('file_selection');
          }}
          optionsMenu={
            <List disablePadding
              subheader={
                <ListSubheader>Call</ListSubheader>
              }
            >
              <ListItem secondaryAction={
                sharedFile && <Tooltip title="Close File">
                  <IconButton
                    {...targets(`exchange close file`, `close the currently shared file`)}
                    color="error"
                    onClick={() => setSharedFile(undefined)}
                    children={<InsertPageBreak />}
                  />
                </Tooltip>
              }>
                <Box sx={classes.darkRounded} mr={1}>
                  {connected && <Tooltip title="Stop Voice or Video" children={
                    <Button
                      {...targets(`exchange leave call`, `leave the voice or video call`)}
                      onClick={() => leaveCall()}
                    >
                      Leave Call
                    </Button>
                  } />}
                  {(!connected || audioOnly) && <Tooltip title="Start Video" children={
                    <IconButton
                      {...targets(`exchange start video call`, `start a video call`)}
                      disabled={'start' !== canStartStop}
                      onClick={() => setLocalStreamAndBroadcast(true)}
                    >
                      <VideocamIcon fontSize="small" />
                    </IconButton>
                  } />}
                  {!connected && <Tooltip title="Start Audio" children={
                    <IconButton
                      {...targets(`exchange start voice call`, `start a voice call`)}
                      disabled={'start' !== canStartStop}
                      onClick={() => setLocalStreamAndBroadcast(false)}
                    >
                      <CallIcon fontSize="small" />
                    </IconButton>
                  } />}
                </Box>
                <Tooltip title="Hide/Show Messages" children={
                  <IconButton
                    {...targets(`exchange toggle chat`, `toggle the chat visbility`)}
                    color="primary"
                    onClick={() => setChatOpen(!chatOpen)}
                  >
                    <ChatBubbleIcon fontSize="small" />
                  </IconButton>
                } />
                <Tooltip title="Open File" children={
                  <IconButton
                    {...targets(`exchange select file`, `select a file to view and share`)}
                    color="primary"
                    onClick={() => setDialog('file_selection')}
                  >
                    <FileCopyIcon fontSize="small" />
                  </IconButton>
                } />
              </ListItem>
            </List>
          }
        />

      </Grid>
    </Grid>
  </>;
}

export default Exchange;
