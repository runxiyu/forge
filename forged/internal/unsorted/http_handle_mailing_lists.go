// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-message"
	"github.com/jackc/pgx/v5"
	"github.com/microcosm-cc/bluemonday"
	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/render"
	"go.lindenii.runxiyu.org/forge/forged/internal/web"
)

// httpHandleMailingListIndex renders the page for a single mailing list.
func (s *Server) httpHandleMailingListIndex(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	groupPath := params["group_path"].([]string)
	listName := params["list_name"].(string)

	groupID, err := s.resolveGroupPath(request.Context(), groupPath)
	if errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error resolving group: "+err.Error())
		return
	}

	var (
		listID    int
		listDesc  string
		emailRows pgx.Rows
		emails    []map[string]any
	)

	if err := s.database.QueryRow(request.Context(),
		`SELECT id, COALESCE(description, '') FROM mailing_lists WHERE group_id = $1 AND name = $2`,
		groupID, listName,
	).Scan(&listID, &listDesc); errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error loading mailing list: "+err.Error())
		return
	}

	emailRows, err = s.database.Query(request.Context(), `SELECT id, title, sender, date FROM mailing_list_emails WHERE list_id = $1 ORDER BY date DESC, id DESC LIMIT 200`, listID)
	if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error loading list emails: "+err.Error())
		return
	}
	defer emailRows.Close()

	for emailRows.Next() {
		var (
			id            int
			title, sender string
			dateVal       time.Time
		)
		if err := emailRows.Scan(&id, &title, &sender, &dateVal); err != nil {
			web.ErrorPage500(s.templates, writer, params, "Error scanning list emails: "+err.Error())
			return
		}
		emails = append(emails, map[string]any{
			"id":     id,
			"title":  title,
			"sender": sender,
			"date":   dateVal,
		})
	}
	if err := emailRows.Err(); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error iterating list emails: "+err.Error())
		return
	}

	params["list_name"] = listName
	params["list_description"] = listDesc
	params["list_emails"] = emails

	listURLRoot := "/"
	segments := params["url_segments"].([]string)
	for _, part := range segments[:params["separator_index"].(int)+3] {
		listURLRoot += part + "/"
	}
	params["list_email_address"] = listURLRoot[1:len(listURLRoot)-1] + "@" + s.config.LMTP.Domain

	var count int
	if err := s.database.QueryRow(request.Context(), `
	SELECT COUNT(*) FROM user_group_roles WHERE user_id = $1 AND group_id = $2
	`, params["user_id"].(int), groupID).Scan(&count); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error checking access: "+err.Error())
		return
	}
	params["direct_access"] = (count > 0)

	s.renderTemplate(writer, "mailing_list", params)
}

// httpHandleMailingListRaw serves a raw email by ID from a list.
func (s *Server) httpHandleMailingListRaw(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	groupPath := params["group_path"].([]string)
	listName := params["list_name"].(string)
	idStr := params["email_id"].(string)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		web.ErrorPage400(s.templates, writer, params, "Invalid email id")
		return
	}

	groupID, err := s.resolveGroupPath(request.Context(), groupPath)
	if errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error resolving group: "+err.Error())
		return
	}

	var listID int
	if err := s.database.QueryRow(request.Context(),
		`SELECT id FROM mailing_lists WHERE group_id = $1 AND name = $2`,
		groupID, listName,
	).Scan(&listID); errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error loading mailing list: "+err.Error())
		return
	}

	var content []byte
	if err := s.database.QueryRow(request.Context(),
		`SELECT content FROM mailing_list_emails WHERE id = $1 AND list_id = $2`, id, listID,
	).Scan(&content); errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error loading email content: "+err.Error())
		return
	}

	writer.Header().Set("Content-Type", "message/rfc822")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(content)
}

// httpHandleMailingListSubscribers lists and manages the subscribers for a mailing list.
func (s *Server) httpHandleMailingListSubscribers(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	groupPath := params["group_path"].([]string)
	listName := params["list_name"].(string)

	groupID, err := s.resolveGroupPath(request.Context(), groupPath)
	if errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error resolving group: "+err.Error())
		return
	}

	var listID int
	if err := s.database.QueryRow(request.Context(),
		`SELECT id FROM mailing_lists WHERE group_id = $1 AND name = $2`,
		groupID, listName,
	).Scan(&listID); errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error loading mailing list: "+err.Error())
		return
	}

	var count int
	if err := s.database.QueryRow(request.Context(), `SELECT COUNT(*) FROM user_group_roles WHERE user_id = $1 AND group_id = $2`, params["user_id"].(int), groupID).Scan(&count); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error checking access: "+err.Error())
		return
	}
	directAccess := (count > 0)
	if request.Method == http.MethodPost {
		if !directAccess {
			web.ErrorPage403(s.templates, writer, params, "You do not have direct access to this list")
			return
		}
		switch request.FormValue("op") {
		case "add":
			email := strings.TrimSpace(request.FormValue("email"))
			if email == "" || !strings.Contains(email, "@") {
				web.ErrorPage400(s.templates, writer, params, "Valid email is required")
				return
			}
			if _, err := s.database.Exec(request.Context(),
				`INSERT INTO mailing_list_subscribers (list_id, email) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				listID, email,
			); err != nil {
				web.ErrorPage500(s.templates, writer, params, "Error adding subscriber: "+err.Error())
				return
			}
			misc.RedirectUnconditionally(writer, request)
			return
		case "remove":
			idStr := request.FormValue("id")
			id, err := strconv.Atoi(idStr)
			if err != nil || id <= 0 {
				web.ErrorPage400(s.templates, writer, params, "Invalid id")
				return
			}
			if _, err := s.database.Exec(request.Context(),
				`DELETE FROM mailing_list_subscribers WHERE id = $1 AND list_id = $2`, id, listID,
			); err != nil {
				web.ErrorPage500(s.templates, writer, params, "Error removing subscriber: "+err.Error())
				return
			}
			misc.RedirectUnconditionally(writer, request)
			return
		default:
			web.ErrorPage400(s.templates, writer, params, "Unknown operation")
			return
		}
	}

	rows, err := s.database.Query(request.Context(), `SELECT id, email FROM mailing_list_subscribers WHERE list_id = $1 ORDER BY email ASC`, listID)
	if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error loading subscribers: "+err.Error())
		return
	}
	defer rows.Close()
	var subs []map[string]any
	for rows.Next() {
		var id int
		var email string
		if err := rows.Scan(&id, &email); err != nil {
			web.ErrorPage500(s.templates, writer, params, "Error scanning subscribers: "+err.Error())
			return
		}
		subs = append(subs, map[string]any{"id": id, "email": email})
	}
	if err := rows.Err(); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error iterating subscribers: "+err.Error())
		return
	}

	params["list_name"] = listName
	params["subscribers"] = subs
	params["direct_access"] = directAccess

	s.renderTemplate(writer, "mailing_list_subscribers", params)
}

// httpHandleMailingListMessage renders a single archived message.
func (s *Server) httpHandleMailingListMessage(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	groupPath := params["group_path"].([]string)
	listName := params["list_name"].(string)
	idStr := params["email_id"].(string)
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		web.ErrorPage400(s.templates, writer, params, "Invalid email id")
		return
	}

	groupID, err := s.resolveGroupPath(request.Context(), groupPath)
	if errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error resolving group: "+err.Error())
		return
	}

	var listID int
	if err := s.database.QueryRow(request.Context(),
		`SELECT id FROM mailing_lists WHERE group_id = $1 AND name = $2`,
		groupID, listName,
	).Scan(&listID); errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error loading mailing list: "+err.Error())
		return
	}

	var raw []byte
	if err := s.database.QueryRow(request.Context(),
		`SELECT content FROM mailing_list_emails WHERE id = $1 AND list_id = $2`, id, listID,
	).Scan(&raw); errors.Is(err, pgx.ErrNoRows) {
		web.ErrorPage404(s.templates, writer, params)
		return
	} else if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error loading email content: "+err.Error())
		return
	}

	entity, err := message.Read(bytes.NewReader(raw))
	if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error parsing email content: "+err.Error())
		return
	}

	subj := entity.Header.Get("Subject")
	from := entity.Header.Get("From")
	dateStr := entity.Header.Get("Date")
	var dateVal time.Time
	if t, err := mail.ParseDate(dateStr); err == nil {
		dateVal = t
	}

	isHTML, body := extractBody(entity)
	var bodyHTML any
	if isHTML {
		bodyHTML = bluemonday.UGCPolicy().SanitizeBytes([]byte(body))
	} else {
		bodyHTML = render.EscapeHTML(body)
	}

	params["email_subject"] = subj
	params["email_from"] = from
	params["email_date_raw"] = dateStr
	params["email_date"] = dateVal
	params["email_body_html"] = bodyHTML

	s.renderTemplate(writer, "mailing_list_message", params)
}

func extractBody(e *message.Entity) (bool, string) {
	ctype := e.Header.Get("Content-Type")
	mtype, params, _ := mime.ParseMediaType(ctype)
	var plain string
	var htmlBody string

	if strings.HasPrefix(mtype, "multipart/") {
		b := params["boundary"]
		if b == "" {
			data, _ := io.ReadAll(e.Body)
			return false, string(data)
		}
		mr := multipart.NewReader(e.Body, b)
		for {
			part, err := mr.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				break
			}
			ptype, _, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
			pdata, _ := io.ReadAll(part)
			switch strings.ToLower(ptype) {
			case "text/plain":
				if plain == "" {
					plain = string(pdata)
				}
			case "text/html":
				if htmlBody == "" {
					htmlBody = string(pdata)
				}
			}
		}
		if plain != "" {
			return false, plain
		}
		if htmlBody != "" {
			return true, htmlBody
		}
		return false, ""
	}

	data, _ := io.ReadAll(e.Body)
	switch strings.ToLower(mtype) {
	case "", "text/plain":
		return false, string(data)
	case "text/html":
		return true, string(data)
	default:
		return false, string(data)
	}
}
