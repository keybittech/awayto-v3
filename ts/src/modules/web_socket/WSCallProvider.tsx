import React, { useRef, useCallback, useMemo, useState, useEffect } from 'react';

import { ExchangeSessionAttributes, SenderStreams, SocketMessage, SocketActions, useComponents, useContexts, useUtil, useWebSocketSubscribe } from 'awayto/hooks';

const {
  REACT_APP_TURN_NAME,
  REACT_APP_TURN_PASS
} = process.env as { [prop: string]: string };

const peerConnectionConfig = {
  'iceServers': [
    { urls: `turn:${location.hostname}:3478`, credential: 'turnpass', username: 'turnuser' },
    { urls: `stun:${location.hostname}:3478` },
  ]
};

declare global {
  interface IComponent {
    topicId?: string;
    topicMessages?: SocketMessage[];
    setTopicMessages?(selector: (prop: SocketMessage[]) => SocketMessage[]): void;
  }
}

export function WSCallProvider({ children, topicId, setTopicMessages }: IComponent): React.JSX.Element {
  if (!topicId) return <>{children}</>;

  const { WSCallContext } = useContexts();
  const { setSnack } = useUtil();
  const { Video } = useComponents();

  const [streamsUpdated, setStreamsUpdated] = useState('');
  const [audioOnly, setAudioOnly] = useState(false);
  const [canStartStop, setCanStartStop] = useState('start');
  const localStream = useRef<MediaStream>();
  const pingInit = useRef(false);
  const callOptionRef = useRef<string[]>([]);
  const senderStreamsRef = useRef<SenderStreams>({});

  const iceCandidateQueue = useRef<{ [prop: string]: RTCIceCandidate[] }>({});

  const setUpSender = (senderId: string) => {
    if (!senderStreamsRef.current[senderId]) {
      const pc = new RTCPeerConnection(peerConnectionConfig);
      iceCandidateQueue.current[senderId] = [];

      if (localStream.current) {
        localStream.current.getTracks().forEach(track => {
          pc?.addTrack(track);
        });
      }

      pc.onicecandidate = event => {


        const iceQueue = iceCandidateQueue.current[senderId];

        if (event.candidate) {
          iceQueue.push(event.candidate);

          const currentSender = senderStreamsRef.current[senderId];

          if (currentSender?.pc && !currentSender.pc.pendingLocalDescription && !currentSender.pc.pendingRemoteDescription && currentSender.pc.currentLocalDescription && currentSender.pc.currentRemoteDescription) {
            sendMessage(SocketActions.RTC, {
              ice: event.candidate,
              target: senderId
            });

          } else {
            iceQueue.push(event.candidate);
            iceCandidateQueue.current[senderId] = iceQueue;
          }
        }
        // When we generate an ICE candidate, send it to the peer
        // if (event.candidate !== null) {
        //   sendMessage('rtc', {
        //     ice: event.candidate,
        //     target: senderId
        //   });
        // }
      };

      pc.ontrack = event => {

        // When receiving a track from a peer, add it to their mediaStream
        const currentSender = senderStreamsRef.current[senderId];
        if (!currentSender?.pc) return;
        console.log('RECEIVING TRACK');
        currentSender.mediaStream = currentSender.mediaStream ?? new MediaStream();
        currentSender.mediaStream.addTrack(event.track);
        senderStreamsRef.current[senderId] = currentSender;
        setStreamsUpdated((new Date()).toString());
      };

      pc.oniceconnectionstatechange = () => {
        // Clean up failed connections
        const currentSender = senderStreamsRef.current[senderId];
        if (currentSender?.pc && ['failed', 'closed', 'disconnected'].includes(currentSender.pc.iceConnectionState)) {
          setStreamsUpdated((new Date()).toString());
          currentSender.mediaStream = undefined;
          currentSender.pc.close();
          delete senderStreamsRef.current[senderId];
        }
      };

      senderStreamsRef.current[senderId] = { pc };
    }
  };

  const {
    subscribed,
    userList,
    connectionId,
    sendMessage
  } = useWebSocketSubscribe<ExchangeSessionAttributes>(topicId, async ({ sender, action, payload }) => {

    const timestamp = (new Date()).toString();
    const { formats, target, sdp, ice, message, style } = payload;

    // If this message isn't from my self or it isn't targeted for me and
    // isn't related to any WebRTC messages
    const isUnrelated = [SocketActions.STREAM_INQUIRY, SocketActions.PING_CHANNEL, SocketActions.STOP_STREAM].includes(action)
    if ((sender === connectionId || (target && target !== connectionId)) || (!target && !isUnrelated)) {
      return;
    }

    if (SocketActions.TEXT === action && setTopicMessages && message && style) {
      for (const user of Object.values(userList)) {
        if (user?.cids.includes(sender)) {
          setTopicMessages(m => [...m, {
            ...user,
            sender,
            style,
            message,
            timestamp
          }]);
        }
      }
    } else if (SocketActions.PING_CHANNEL === action && !!localStream.current) {
      // When new chatters ping the channel, and we're already streaming,
      // initate setup
      sendMessage(SocketActions.START_STREAM, { target: sender });
    } else if (SocketActions.STOP_STREAM === action) {
      // Only remove the member's media stream when they stop streaming
      // i.e. continue to allow our stream to flow to them
      senderStreamsRef.current[sender].mediaStream = undefined;
      setStreamsUpdated(timestamp);
    } else if (SocketActions.STREAM_INQUIRY === action && !senderStreamsRef.current[sender]) {
      // Add approval step
      sendMessage(SocketActions.START_STREAM, { target: sender });
    } else if (SocketActions.START_STREAM === action) {
      // Parties to an incoming caller's 'start-stream' will see this, and 
      // then notify the caller that they exist in return
      // The caller gets a party member's 'peer-response', and sets them 
      // up in return
      // if (setTopicMessages) {
      //   for (const user of userList.values()) {
      //     if (user.cids.includes(sender)) {
      //       setTopicMessages(m => [...m, {
      //         ...user,
      //         sender,
      //         style: 'utterance',
      //         message: `Joined call${formats ? ' with ' + (formats.indexOf('video') > -1 ? 'video' : 'voice') : ''}.`,
      //         timestamp
      //       }]);
      //     }
      //   }
      // }

      setUpSender(sender);

      const startedSender = senderStreamsRef.current[sender];

      // const startedSender = senderStreamsRef.current[sender] || { peerResponse: 'peer-response' === action };
      // startedSender.pc = startedSender.pc || new RTCPeerConnection(peerConnectionConfig);
      //
      // const sentTracks = startedSender.pc?.getSenders().map(ts => ts.track?.id);

      // If we already sent tracks to the pc, we know we don't need to setup
      // an entirely new pc for them -- and their client will create an offer
      // anyway upon starting their stream
      // if (sentTracks.length >= 1) return;

      // startedSender.pc.onicecandidate = event => {
      //   // When we generate an ICE candidate, send it to the peer
      //   if (event.candidate !== null) {
      //     sendMessage('rtc', {
      //       ice: event.candidate,
      //       target: sender
      //     });
      //   }
      // };
      //
      // startedSender.pc.ontrack = event => {
      //   // When receiving a track from a peer, add it to their mediaStream
      //   const currentSender = senderStreamsRef.current[sender];
      //   if (!currentSender?.pc) return;
      //   currentSender.mediaStream = currentSender.mediaStream ?? new MediaStream();
      //   currentSender.mediaStream.addTrack(event.track);
      //   senderStreamsRef.current[sender] = currentSender;
      //   setStreamsUpdated((new Date()).toString());
      // };
      //
      // startedSender.pc.oniceconnectionstatechange = () => {
      //   // Clean up failed connections
      //   const currentSender = senderStreamsRef.current[sender];
      //   if (currentSender?.pc && ['failed', 'closed', 'disconnected'].includes(currentSender.pc.iceConnectionState)) {
      //     setStreamsUpdated((new Date()).toString());
      //     currentSender.mediaStream = undefined;
      //     currentSender.pc.close();
      //     delete senderStreamsRef.current[sender];
      //   }
      // };

      // if (localStream.current) {
      //   // If this client is currently streaming when setting up a peer,
      //   // include the existing tracks
      //   const tracks = localStream.current.getTracks();
      //   const sentTracks = startedSender.pc?.getSenders().map(ts => ts.track?.id);
      //   tracks.filter(t => !sentTracks.includes(t.id)).forEach(track => {
      //     startedSender.pc?.addTrack(track)
      //   });
      // }

      // if (startedSender.peerResponse) {
      // In a situation where no one is currently streaming, User A sends
      // 'start-stream', that will be received here in the "else" block. In
      // response, User B sends a 'peer-response' to User A, caught here,
      // and begins the WebRTC transaction by sending an offer.

      if (startedSender?.pc) {

        const description = await startedSender.pc.createOffer();
        await startedSender.pc.setLocalDescription(description);

        sendMessage(SocketActions.RTC, {
          sdp: startedSender.pc.localDescription,
          target: sender
        });

      }
      // } else {
      //   sendMessage('peer-response', {
      //     target: sender
      //   });
      // }

      senderStreamsRef.current[sender] = startedSender;

    } else if (sdp) {
      console.log(sdp)
      // Standard WebRTC SDP message handling
      setUpSender(sender);
      const currentSender = senderStreamsRef.current[sender];
      if (currentSender?.pc) {

        if (currentSender.pc.signalingState === 'stable') {
          await currentSender.pc.setLocalDescription(undefined);
        }

        await currentSender.pc.setRemoteDescription(new RTCSessionDescription(sdp));

        if ('offer' === sdp.type) {
          const desc = await currentSender.pc.createAnswer();
          await currentSender.pc.setLocalDescription(desc);
          sendMessage(SocketActions.RTC, {
            sdp: currentSender.pc.localDescription,
            target: sender
          });
        } else {
          for (const candidate of iceCandidateQueue.current[sender]) {
            sendMessage(SocketActions.RTC, {
              ice: candidate,
              target: sender
            });
          }

          iceCandidateQueue.current[sender] = [];
        }

        senderStreamsRef.current[sender] = currentSender;
      }
    } else if (ice) {
      console.log(ice)
      // Standard WebRTC ICE message handling
      const currentSender = senderStreamsRef.current[sender];
      if (currentSender?.pc && !currentSender.pc.pendingLocalDescription && !currentSender.pc.pendingRemoteDescription && currentSender.pc.currentLocalDescription && currentSender.pc.currentRemoteDescription && !['failed', 'closed', 'disconnected'].includes(currentSender.pc.iceConnectionState)) {
        await currentSender.pc.addIceCandidate(new RTCIceCandidate(ice));
        senderStreamsRef.current[sender] = currentSender;
      }
    }
  });

  const setLocalStreamAndBroadcast = useCallback((video: boolean): void => {
    async function go() {
      console.log({ localStream, canStartStop })
      try {
        if (!localStream.current && 'start' === canStartStop) {
          setCanStartStop('');
          setAudioOnly(!video);

          const callOptions: MediaStreamConstraints = {
            audio: {
              echoCancellation: true,
              noiseSuppression: true,
              autoGainControl: true,
              latency: 0.01,
              channelCount: 2,
              sampleRate: 48000,
              sampleSize: 16
            }
          };

          if (video) {
            callOptions.video = {
              width: 520,
              height: 390,
              frameRate: { max: 30 }
            };
          }

          callOptionRef.current = Object.keys(callOptions);

          localStream.current = await navigator.mediaDevices.getUserMedia(callOptions);


          if (Object.keys(senderStreamsRef.current).length) {
            const tracks = localStream.current.getTracks();

            // Handle ongoing pc connections by sending a new offer with the new
            // media tracks
            for (const senderId in senderStreamsRef.current) {
              const sender = senderStreamsRef.current[senderId];

              tracks.forEach(track => sender.pc?.addTrack(track));

              const description = await sender.pc?.createOffer();
              await sender.pc?.setLocalDescription(description);

              senderStreamsRef.current[senderId] = sender;

              sendMessage(SocketActions.RTC, {
                sdp: sender.pc?.localDescription,
                target: senderId
              });
            }
          }

          // trackStream(mediaStream); -- TODO: Check support for this in browsers some day

          sendMessage(SocketActions.STREAM_INQUIRY);
          setCanStartStop('stop');
        }
      } catch (error) {
        setCanStartStop('start');
        setSnack({ snackOn: (error as DOMException).message, snackType: 'error' });
      }
    }
    void go();
  }, [canStartStop]);

  const leaveCall = () => {
    if (localStream.current) {
      localStream.current.getTracks().forEach(t => {
        localStream.current?.removeTrack(t);
        t.stop();
      });
      localStream.current = undefined;
    }

    sendMessage(SocketActions.STOP_STREAM);
    setCanStartStop('start');

    // speechRecognizer.current?.stop();
    // speechRecognizer.current = undefined;
  }

  const senderStreamsElements = useMemo(() => {
    const streams = Object.keys(senderStreamsRef.current).map(sender => senderStreamsRef.current[sender].mediaStream ?
      <Video key={sender} autoPlay srcObject={senderStreamsRef.current[sender].mediaStream} /> :
      undefined
    );
    return streams.filter(s => !!s).length ? streams : undefined;
  }, [streamsUpdated]);

  const localStreamElement = useMemo(() => {
    return canStartStop && localStream.current ? <Video key={'local-video'} autoPlay controls srcObject={localStream.current} /> : undefined
  }, [canStartStop]);

  useEffect(() => {
    // Only run once we are subscribed and we haven't yet pinged the channel
    if (subscribed && !pingInit.current) {
      pingInit.current = true;
      if (Object.keys(userList).length >= 2) sendMessage(SocketActions.PING_CHANNEL);

      // When we leave the page, stop all ongoing peer connections and reset the
      // streams container
      return () => {
        for (const senderId in senderStreamsRef.current) {
          const senderStream = senderStreamsRef.current[senderId];

          if (senderStream && senderStream.pc) {
            senderStream.pc.close();
          }
        }
        senderStreamsRef.current = {};
      }

    }
  }, [subscribed]);

  const wsTextContext = {
    audioOnly,
    connected: !!localStream.current,
    canStartStop,
    setLocalStreamAndBroadcast,
    leaveCall,
    senderStreamsElements,
    localStreamElement
  } as WSCallContextType | null;

  return useMemo(() => !WSCallContext ? <></> :
    <WSCallContext.Provider value={wsTextContext}>
      {children}
    </WSCallContext.Provider>,
    [WSCallContext, wsTextContext]
  );
}

export default WSCallProvider;
