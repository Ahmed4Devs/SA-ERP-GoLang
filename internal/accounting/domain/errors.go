package domain

import "errors"

var (
	ErrUnbalancedJournalEntry    = errors.New("unbalanced_journal_entry: total debits must equal total credits")
	ErrInvalidLineAmounts        = errors.New("invalid_line_amounts: each line must contain debit XOR credit greater than zero")
	ErrPeriodNotOpen             = errors.New("period_not_open: fiscal period is closed or locked for posting")
	ErrAccountNotPostable        = errors.New("account_not_postable: account is a parent category and does not allow direct posting")
	ErrAccountInactive           = errors.New("account_inactive: account is inactive")
	ErrAccountNotFound           = errors.New("account_not_found: account does not exist")
	ErrJournalNotFound           = errors.New("journal_not_found: journal subledger does not exist")
	ErrFiscalPeriodNotFound      = errors.New("fiscal_period_not_found: no active fiscal period matches the accounting date")
	ErrJournalEntryNotFound      = errors.New("journal_entry_not_found: journal entry does not exist")
	ErrCannotModifyPostedEntry   = errors.New("cannot_modify_posted_entry: posted or reversed entries cannot be edited or deleted")
	ErrAlreadyPosted             = errors.New("already_posted: journal entry is already posted")
	ErrAlreadyReversed           = errors.New("already_reversed: journal entry is already reversed")
	ErrCannotReverseUnposted     = errors.New("cannot_reverse_unposted: only posted journal entries can be reversed")
	ErrNoLinesProvided           = errors.New("no_lines_provided: journal entry must contain at least two balanced lines")
	ErrDuplicateLineNumber       = errors.New("duplicate_line_number: journal lines must have unique ascending numbers")
)
