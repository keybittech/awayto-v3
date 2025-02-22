import React, { useState, useEffect } from 'react';
import Avatar from '@mui/material/Avatar';
import { useFileStore } from 'awayto/hooks';

interface AsyncAvatarProps extends IComponent {
  image: string;
}

export function AsyncAvatar({ image }: AsyncAvatarProps): React.JSX.Element {
  const [url, setUrl] = useState('');
  const fileStore = useFileStore();

  useEffect(() => {
    async function getImage() {
      if (fileStore && image) {
        setUrl(await fileStore?.get(image));
      }
    }
    void getImage();
  }, [fileStore, image])

  return url.length ? <Avatar src={url || ''} /> : <></>
}

export default AsyncAvatar;
