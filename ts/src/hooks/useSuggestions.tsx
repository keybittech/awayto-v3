import React, { useState, useCallback, useMemo } from 'react';

import Link from '@mui/material/Link';

import { nid, obfuscate } from './util';
import { siteApi } from './api';
import { IPrompts } from './assist';

type SuggestFn = (props: { id: IPrompts, prompt: string }) => void;
type SuggestionsComp = IComponent & {
  staticSuggestions: string;
  handleSuggestion: (val: string) => void;
  hideSuggestions?: boolean;
};

export function useSuggestions(refName: string): {
  suggestions: string[];
  suggest: SuggestFn;
  comp(props: SuggestionsComp): React.JSX.Element;
} {

  const [suggestions, setSuggestions] = useState(JSON.parse(localStorage.getItem(refName + '_suggestions') || '[]') as string[]);
  const [history, setHistory] = useState(JSON.parse(localStorage.getItem(refName + '_suggestion_history') || '{}') as Record<string, string[]>);

  const { data: userProfileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();
  const [getSuggestion] = siteApi.useLazyAssistServiceGetSuggestionQuery();

  const userGroup = useMemo(() => {
    return Object.values(userProfileRequest?.userProfile.groups || {}).find(g => g.active) || {}
  }, [userProfileRequest?.userProfile.groups]);
  const allowSuggestions = userGroup.ai;

  const suggest: SuggestFn = useCallback(({ id, prompt }) => {
    try {
      if (!allowSuggestions) {
        return
      }

      const promptKey = obfuscate(prompt);

      if (history.hasOwnProperty(promptKey)) {
        setSuggestions(history[promptKey]);
        localStorage.setItem(refName + '_suggestions', JSON.stringify(history[promptKey]));
        return;
      }

      getSuggestion({ id: id.toString(), prompt }).unwrap().then(({ promptResult }) => {
        if (promptResult.length > 0) {
          setSuggestions(promptResult);
          const newHistory = { ...history, [promptKey]: promptResult };
          setHistory(newHistory);
          localStorage.setItem(refName + '_suggestions', JSON.stringify(promptResult));
          localStorage.setItem(refName + '_suggestion_history', JSON.stringify(newHistory));
        }
      }).catch(console.error);
    } catch (error) {
      console.log(error);
    }
  }, [allowSuggestions, history]);

  const comp = useCallback(({ staticSuggestions, handleSuggestion, hideSuggestions }: SuggestionsComp) => {
    const compId = nid('v4');
    return suggestions.length && allowSuggestions && !hideSuggestions ? <>
      AI: {suggestions.filter(s => s.toLowerCase() !== 'admin').map((suggestion, i) => {
        return <span key={`${compId}-selection-${i}`}>
          <Link sx={{ cursor: 'pointer' }} onClick={() => {
            handleSuggestion(suggestion);
          }}>
            {suggestion}
          </Link>{i !== suggestions.length - 1 ? ',' : ''}&nbsp;
        </span>
      })}
    </> : <>{staticSuggestions}</>;
  }, [allowSuggestions, suggestions]);

  return { suggestions, suggest, comp };
}

export default useSuggestions;
