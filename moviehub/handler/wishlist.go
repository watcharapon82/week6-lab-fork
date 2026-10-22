package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"moviehub/mw"
	"moviehub/store"
)

type wishlistReq struct {
	Note *string `json:"note"`
}

func (h *Handler) MyWishlist(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	items, err := h.Store.ListWishlist(r.Context(), who.ID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) PutWishlistItem(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "movieId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	var req wishlistReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	if req.Note != nil {
		note := strings.TrimSpace(*req.Note)
		if utf8.RuneCountInString(note) > 200 {
			writeError(w, http.StatusBadRequest, "BAD_NOTE", "โน้ตต้องยาวไม่เกิน 200 ตัวอักษร")
			return
		}
		req.Note = &note
	}
	who := mw.WhoFrom(r.Context())
	item, err := h.Store.PutWishlistItem(r.Context(), who.ID, movieID, req.Note)
	if errors.Is(err, store.ErrNoMovie) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteWishlistItem(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "movieId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.DeleteWishlistItem(r.Context(), who.ID, movieID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "WISHLIST_ITEM_NOT_FOUND", "ไม่พบหนังในรายการอยากดู")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
