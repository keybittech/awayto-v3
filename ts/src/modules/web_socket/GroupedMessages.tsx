import React from 'react';

import Grid from '@mui/material/Grid';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import Chip from '@mui/material/Chip';
import Avatar from '@mui/material/Avatar';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';

import RecordVoiceOverIcon from '@mui/icons-material/RecordVoiceOver';
import TextFieldsIcon from '@mui/icons-material/TextFields';

import { SocketMessage, utcDTLocal } from 'awayto/hooks';
import capitalize from '@mui/material/utils/capitalize';

interface IComponent {
  topicMessages?: SocketMessage[];
  action?: () => void;
}

interface GroupedMessage extends Omit<SocketMessage, 'message'> {
  messages: string[];
}

function GroupedMessages({ topicMessages: messages }: IComponent): React.JSX.Element {
  if (!messages) return <></>;

  const groupedMessages: GroupedMessage[] = [];
  let currentGroup: GroupedMessage | null = null;
  messages.forEach((msg, i) => {
    if (i === 0 || messages[i - 1].scid !== msg.scid || messages[i - 1].style !== msg.style) {
      currentGroup = {
        ...msg,
        messages: [msg.message]
      };
      groupedMessages.push(currentGroup);
    } else if (currentGroup) {
      currentGroup.messages.push(msg.message);
    }
  });

  return <>
    {groupedMessages.map((group, i) => {
      const Action = group.action;
      return <Card sx={{ marginBottom: '8px' }} key={`${group.scid}_group_${i}`}>
        <CardContent>
          <Grid container spacing={1}>
            <Grid>
              <Avatar sx={{ color: 'black', backgroundColor: group.color, fontStyle: 'bold' }}>{group.name}</Avatar>
            </Grid>
            <Grid sx={{ flex: 1 }}>
              <Chip size="small" variant="outlined" label={group.role} />
              <Grid>
                <Typography variant="caption">{utcDTLocal(group.timestamp || '')}</Typography>
              </Grid>
            </Grid>
            <Grid>
              <Tooltip title={capitalize(group.style)}>
                <IconButton disableRipple>
                  {'utterance' == group.style && <RecordVoiceOverIcon />}
                  {'written' == group.style && <TextFieldsIcon />}
                </IconButton>
              </Tooltip>
            </Grid>
          </Grid>
          {group.messages.map((message, j) => (
            <Typography color="primary" style={{ overflowWrap: 'anywhere', whiteSpace: 'pre-wrap' }} key={`${group.scid}_msg_${j}`}>
              {message}
            </Typography>
          ))}
          {Action && <Action />}
        </CardContent>
      </Card>
    })}
  </>;
}

export default GroupedMessages;
