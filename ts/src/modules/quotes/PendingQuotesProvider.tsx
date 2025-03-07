import React, { useMemo, useState } from 'react';

import { siteApi, useUtil, plural, bookingFormat, IBooking, IQuote } from 'awayto/hooks';

import PendingQuotesContext, { PendingQuotesContextType } from './PendingQuotesContext';

export function PendingQuotesProvider({ children }: IComponent): React.JSX.Element {

  const { setSnack, openConfirm } = useUtil();
  const [disableQuote] = siteApi.useQuoteServiceDisableQuoteMutation();
  const [postBooking] = siteApi.useBookingServicePostBookingMutation();

  const [pendingQuotesChanged, setPendingQuotesChanged] = useState(false);
  const [selectedPendingQuotes, setSelectedPendingQuotes] = useState<string[]>([]);

  const { data: profileRequest, refetch: getUserProfileDetails } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const pendingQuotes = useMemo(() => Object.values(profileRequest?.userProfile?.quotes || {}) as Required<IQuote>[], [profileRequest?.userProfile]);

  const pendingQuotesContext = {
    pendingQuotes,
    pendingQuotesChanged,
    selectedPendingQuotes,
    setSelectedPendingQuotes,
    handleSelectPendingQuote(quote) {
      const currentIndex = selectedPendingQuotes.indexOf(quote);
      const newChecked = [...selectedPendingQuotes];

      if (currentIndex === -1) {
        newChecked.push(quote);
      } else {
        newChecked.splice(currentIndex, 1);
      }

      setSelectedPendingQuotes(newChecked);
    },
    handleSelectPendingQuoteAll() {
      const pendingQuotesSet = selectedPendingQuotes.length === pendingQuotes.length ?
        selectedPendingQuotes.filter(v => !pendingQuotes.map(pq => pq.id).includes(v)) :
        [...selectedPendingQuotes, ...pendingQuotes.filter(v => !selectedPendingQuotes.includes(v.id)).map(pq => pq.id)];

      setSelectedPendingQuotes(pendingQuotesSet);
    },
    approvePendingQuotes() {
      const selectedValues = pendingQuotes.filter(pq => selectedPendingQuotes.includes(pq.id));
      if (!selectedValues.every(s => s.slotDate === selectedValues[0].slotDate && s.scheduleBracketSlotId === selectedValues[0].scheduleBracketSlotId)) {
        setSnack({ snackType: 'error', snackOn: 'Only appointments of the same date and time can be mass approved.' });
        return;
      }

      const { slotDate, startTime, scheduleBracketSlotId } = selectedValues[0];

      const copies = pendingQuotes.filter(q => !selectedValues.some(s => s.id === q.id)).filter(q => q.slotDate === slotDate && q.scheduleBracketSlotId === scheduleBracketSlotId);

      openConfirm({
        isConfirming: true,
        confirmEffect: `Approve ${plural(selectedValues.length, 'request', 'requests')}, creating ${plural(selectedValues.length, 'booking', 'bookings')}, for ${bookingFormat(slotDate, startTime)}.`,
        confirmSideEffect: !copies.length ? undefined : {
          approvalAction: 'Auto-Deny Remaining',
          approvalEffect: `Automatically deny all other requests for ${bookingFormat(slotDate, startTime)} (this cannot be undone).`,
          rejectionAction: 'Confirm Quote/Booking Only',
          rejectionEffect: 'Just submit the approvals.',
        },
        confirmAction: approval => {
          const newBookings = selectedValues.map(s => ({
            quote: s
          }) as IBooking);

          postBooking({ postBookingRequest: { bookings: newBookings } }).unwrap().then(() => {
            const disableQuoteIds = selectedValues.concat(approval ? copies : []).filter(c => !newBookings.some(b => b.id === c.id)).map(s => s.id).join(',');

            disableQuote({ ids: disableQuoteIds }).unwrap().then(() => {
              setSelectedPendingQuotes([]);
              setPendingQuotesChanged(!pendingQuotesChanged);
              void getUserProfileDetails();
            }).catch(console.error);
          }).catch(console.error);
        }
      });
    },
    denyPendingQuotes() {
      disableQuote({ ids: selectedPendingQuotes.join(',') }).unwrap().then(() => {
        setSelectedPendingQuotes([]);
        setPendingQuotesChanged(!pendingQuotesChanged);
        void getUserProfileDetails();
      }).catch(console.error);
    }
  } as PendingQuotesContextType | null;

  return useMemo(() => <PendingQuotesContext.Provider value={pendingQuotesContext}>
    {children}
  </PendingQuotesContext.Provider>, [pendingQuotesContext]);
}

export default PendingQuotesProvider;
