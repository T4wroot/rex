package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// NativeFileReq holds structured requests for direct file syscalls
type NativeFileReq struct {
	Action  string `json:"action"`            // "read", "write", "stat", "list", "delete"
	Path    string `json:"path"`              // target file path
	Content []byte `json:"content,omitempty"` // for write
	Mode    uint32 `json:"mode,omitempty"`    // file permissions
}

// NativeFileRes holds response for file syscalls
type NativeFileRes struct {
	Status  string      `json:"status"` // "ok" or "error"
	Error   string      `json:"error,omitempty"`
	Content []byte      `json:"content,omitempty"`
	Files   []FileInfo  `json:"files,omitempty"`
	Stat    *FileStat   `json:"stat,omitempty"`
}

type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

type FileStat struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

// HandleNativeFileOp executes direct OS syscalls for file management with security checks
func HandleNativeFileOp(payload []byte, al *Allowlist) []byte {
	var req NativeFileReq
	if err := json.Unmarshal(payload, &req); err != nil {
		return makeErrRes(fmt.Sprintf("invalid payload: %v", err))
	}

	if al != nil {
		if req.Action == "write" || req.Action == "delete" {
			cmdStr := fmt.Sprintf("%s %s", req.Action, req.Path)
			allowed, reason := al.IsCommandAllowed(cmdStr)
			if !allowed {
				return makeErrRes(fmt.Sprintf("security policy denied %s: %s", req.Action, reason))
			}
		} else if !al.IsPathAllowed(req.Path) && al.Mode == LevelAllowlist {
			return makeErrRes(fmt.Sprintf("access to path %q denied by allowlist policy", req.Path))
		}
	}

	switch req.Action {
	case "read":
		data, err := os.ReadFile(req.Path)
		if err != nil {
			return makeErrRes(err.Error())
		}
		return makeOkRes(NativeFileRes{Status: "ok", Content: data})

	case "write":
		perm := os.FileMode(0644)
		if req.Mode > 0 {
			perm = os.FileMode(req.Mode)
		}
		if err := os.WriteFile(req.Path, req.Content, perm); err != nil {
			return makeErrRes(err.Error())
		}
		return makeOkRes(NativeFileRes{Status: "ok"})

	case "stat":
		st, err := os.Stat(req.Path)
		if err != nil {
			return makeErrRes(err.Error())
		}
		return makeOkRes(NativeFileRes{
			Status: "ok",
			Stat: &FileStat{
				Name:    st.Name(),
				Size:    st.Size(),
				Mode:    st.Mode().String(),
				ModTime: st.ModTime().Format("2006-01-02 15:04:05"),
				IsDir:   st.IsDir(),
			},
		})

	case "list":
		entries, err := os.ReadDir(req.Path)
		if err != nil {
			return makeErrRes(err.Error())
		}
		var files []FileInfo
		for _, e := range entries {
			info, _ := e.Info()
			sz := int64(0)
			if info != nil {
				sz = info.Size()
			}
			files = append(files, FileInfo{
				Name:  e.Name(),
				IsDir: e.IsDir(),
				Size:  sz,
			})
		}
		return makeOkRes(NativeFileRes{Status: "ok", Files: files})

	case "delete":
		if err := os.RemoveAll(req.Path); err != nil {
			return makeErrRes(err.Error())
		}
		return makeOkRes(NativeFileRes{Status: "ok"})

	default:
		return makeErrRes(fmt.Sprintf("unknown action: %s", req.Action))
	}
}

func makeErrRes(msg string) []byte {
	b, _ := json.Marshal(NativeFileRes{Status: "error", Error: msg})
	return b
}

func makeOkRes(res NativeFileRes) []byte {
	b, _ := json.Marshal(res)
	return b
}
