import { useMemo, useState } from 'react';

import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';

import Send from '@mui/icons-material/Send';
import { targets } from 'awayto/hooks';

interface SubmitMessageFormProps extends IComponent {
  sendTextMessage?: (msg: string) => void;
}

export function SubmitMessageForm({ sendTextMessage }: SubmitMessageFormProps): React.JSX.Element {

  const [textMessage, setTextMessage] = useState('');

  return useMemo(() => {
    return <form onSubmit={e => {
      e.preventDefault();
      sendTextMessage && sendTextMessage(textMessage);
      setTextMessage('');
    }}>
      <TextField
        {...targets(`submit message form message`, `Type here then press enter...`, `input a message to be sent to the chat`)}
        fullWidth
        multiline
        value={textMessage}
        onChange={e => setTextMessage(e.target.value)}
        slotProps={{
          input: {
            sx: {
              'textarea': {
                overflow: 'auto !important',
                maxHeight: '60px'
              }
            },
            onKeyDown: e => {
              if ('Enter' === e.key && !e.shiftKey) {
                e.preventDefault();
                sendTextMessage && sendTextMessage(textMessage);
                setTextMessage('');
              }
            },
            endAdornment: <InputAdornment position="end">
              <IconButton
                {...targets(`submit message form submit`, `submit the input message to the chat`)}
                type="submit"
              >
                <Send />
              </IconButton>
            </InputAdornment>
          }
        }}
      />
    </form>
  }, [textMessage])
}

export default SubmitMessageForm;
