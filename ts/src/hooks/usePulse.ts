import { useEffect, useRef } from 'react';

// Token active 5 mins, pulse 4 mins
const PULSE_RATE = 4 * 60 * 1000;

export const usePulse = () => {
  const pulseRef = useRef<number | null>(null);
  useEffect(() => {
    const ping = () => {
      fetch('/auth/status').then(res => {
        if (401 == res.status) {
          window.location.reload();
        }
      }).catch(console.error);
    };

    ping();

    pulseRef.current = setInterval(ping, PULSE_RATE);

    return () => {
      if (pulseRef.current) clearInterval(pulseRef.current);
    }
  }, []);
}
