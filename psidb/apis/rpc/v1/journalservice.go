package rpcv1

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/core"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
)

type JournalService struct {
	core *core.Core
}

func NewJournalService(core *core.Core) *JournalService {
	return &JournalService{core: core}
}

type GetHeadRequest struct {
	Dummy bool `json:"dummy"`
}

type GetHeadResponse struct {
	LastRecordID uint64 `json:"last_record_id"`
}

func (ns *JournalService) GetHead(ctx context.Context, req *GetHeadRequest) (res *GetHeadResponse, err error) {
	head, err := ns.core.Journal().GetHead()

	if err != nil {
		return nil, err
	}

	res = &GetHeadResponse{LastRecordID: head}

	return
}

type GetEntryRequest struct {
	RecordID uint64 `json:"record_id"`
}

type GetEntryResponse struct {
	Entry coreapi.JournalEntry `json:"entry"`
}

func (ns *JournalService) GetEntry(ctx context.Context, req *GetEntryRequest) (res *GetEntryResponse, err error) {
	res = &GetEntryResponse{}

	_, err = ns.core.Journal().Read(req.RecordID, &res.Entry)

	if err != nil {
		return nil, err
	}

	return
}

type GetEntryRangeRequest struct {
	From  uint64 `json:"from"`
	Count int    `json:"count"`
}

type GetEntryRangeResponse struct {
	Entry []coreapi.JournalEntry `json:"entries"`
}

func (ns *JournalService) GetEntryRange(ctx context.Context, req *GetEntryRangeRequest) (res *GetEntryRangeResponse, err error) {
	res = &GetEntryRangeResponse{}

	it := ns.core.Journal().Iterate(req.From, req.Count)

	res.Entry = iterators.ToSlice(it)

	return
}
