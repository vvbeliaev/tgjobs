package core

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/pocketbase/pocketbase/core"
)

// Job is the aggregate root for job vacancy domain.
// It wraps a PocketBase record and provides state machine methods.
type Job struct {
	record *core.Record
}

// NewJob creates a Job aggregate from an existing PocketBase record.
func NewJob(record *core.Record) *Job {
	return &Job{record: record}
}

// NewRawJob creates a new Job in raw state from input data.
// The record must be created with the jobs collection.
func NewRawJob(collection *core.Collection, input RawJobInput) *Job {
	record := core.NewRecord(collection)

	// Extract preliminary title from first line
	title := "Pending Analysis"
	lines := strings.Split(input.OriginalText, "\n")
	if len(lines) > 0 && len(lines[0]) > 0 {
		firstLine := lines[0]
		if utf8.RuneCountInString(firstLine) > 100 {
			runes := []rune(firstLine)
			title = string(runes[:97]) + "..."
		} else {
			title = firstLine
		}
	}

	record.Set("title", title)
	record.Set("originalText", input.OriginalText)
	record.Set("channelId", fmt.Sprintf("%d", input.ChannelID))
	record.Set("messageId", input.MessageID)
	record.Set("hash", input.Hash)
	record.Set("raw", input.RawData)
	record.Set("status", string(StatusRaw))

	url := fmt.Sprintf("https://t.me/c/%d/%d", input.ChannelID, input.MessageID)
	record.Set("url", url)

	return &Job{record: record}
}

// --- Getters ---

// ID returns the job's unique identifier.
func (j *Job) ID() string {
	return j.record.Id
}

// Status returns the current job status.
func (j *Job) Status() JobStatus {
	return JobStatus(j.record.GetString("status"))
}

// OriginalText returns the raw job posting text.
func (j *Job) OriginalText() string {
	return j.record.GetString("originalText")
}

// Description returns the processed job description.
func (j *Job) Description() string {
	return j.record.GetString("description")
}

// Record returns the underlying PocketBase record.
// Use this when you need to persist changes via app.Save().
func (j *Job) Record() *core.Record {
	return j.record
}

// --- State Machine Methods ---

// MarkProcessing transitions job from raw to processing state.
func (j *Job) MarkProcessing() error {
	if j.Status() != StatusRaw {
		return errors.New("can only mark processing from raw state")
	}
	j.record.Set("status", string(StatusProcessing))
	return nil
}

// Complete transitions job from processing to processed state with parsed data.
func (j *Job) Complete(data ParsedData) error {
	if j.Status() != StatusProcessing {
		return errors.New("can only complete from processing state")
	}

	j.record.Set("title", data.Title)
	j.record.Set("company", data.Company)
	j.record.Set("salaryMin", data.SalaryMin)
	j.record.Set("salaryMax", data.SalaryMax)
	j.record.Set("currency", data.Currency)
	j.record.Set("grade", data.Grade)
	j.record.Set("location", data.Location)
	j.record.Set("isRemote", data.IsRemote)
	j.record.Set("description", data.Description)
	j.record.Set("skills", data.Skills)
	j.record.Set("status", string(StatusProcessed))

	return nil
}

// Reject transitions job from processing to rejected state.
func (j *Job) Reject(reason string) error {
	if j.Status() != StatusProcessing {
		return errors.New("can only reject from processing state")
	}
	j.record.Set("status", string(StatusRejected))
	// Optionally store rejection reason if field exists
	return nil
}

// IsVacancy checks if the job was determined to be a real vacancy.
func (j *Job) IsVacancy() bool {
	return j.Status() == StatusProcessed
}
