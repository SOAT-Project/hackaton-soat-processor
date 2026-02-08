package domain

import "time"

type VideoProcess struct {
	ProcessID   string
	VideoBucket string
	VideoKey    string
	CreatedAt   time.Time
}

type ProcessResult struct {
	ProcessID  string
	FileBucket string
	FileKey    string
	Success    bool
	Error      error
}

func (r *ProcessResult) ToSuccessMessage() map[string]interface{} {
	return map[string]interface{}{
		"process_id":  r.ProcessID,
		"file_bucket": r.FileBucket,
		"file_key":    r.FileKey,
	}
}

func (r *ProcessResult) ToErrorMessage() map[string]interface{} {
	errorMsg := "unknown error"
	if r.Error != nil {
		errorMsg = r.Error.Error()
	}
	return map[string]interface{}{
		"process_id":    r.ProcessID,
		"error_message": errorMsg,
	}
}
