package dataprocessor

import (
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	dictloader "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/dict_loader"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/events_proto"
	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/cosmos/gogoproto/proto"
	"github.com/klauspost/compress/zstd"
)

// EventFormat represents the format type of the returned data
type EventFormat int

const (
	NativeFormat EventFormat = iota
	CompressedFormat
)

var dictBytes = dictloader.LoadDict()
var zstdDict = zstd.WithEncoderDict(dictBytes)
var zstdLvl = zstd.WithEncoderLevel(zstd.SpeedBestCompression)
var zstdWriter *zstd.Encoder

func init() {
	var err error
	zstdWriter, err = zstd.NewWriter(nil, zstdDict, zstdLvl)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to initialize zstd writer")
	}
}

// EventResult holds the result of EventSolver with type discrimination
type EventResult struct {
	Format         EventFormat
	NativeEvents   []s.Event // populated when Format == NativeFormat
	CompressedData []byte    // populated when Format == CompressedFormat
}

// IsNative returns true if the result contains native format data
func (er *EventResult) IsNative() bool {
	return er.Format == NativeFormat
}

// IsCompressed returns true if the result contains compressed format data
func (er *EventResult) IsCompressed() bool {
	return er.Format == CompressedFormat
}

// GetNativeEvents returns the native events if available, nil otherwise
func (er *EventResult) GetNativeEvents() []s.Event {
	if er.Format == NativeFormat {
		return er.NativeEvents
	}
	return nil
}

// GetCompressedData returns the compressed data if available, nil otherwise
func (er *EventResult) GetCompressedData() []byte {
	if er.Format == CompressedFormat {
		return er.CompressedData
	}
	return nil
}

// EventSolver is a function that solves the event of a transaction
// it will solve the event of a transaction and return the event
//
// It can return data in two formats:
// 1. Native postgres format ([]s.Event)
// 2. Compressed protobuf format ([]byte)
//
// Parameters:
//   - txResponse: a transaction response
//   - useCompressed: if true, returns compressed format; otherwise native format
//
// Returns:
//   - *EventResult: contains either native events or compressed data
//   - error: an error if the event solving fails
func EventSolver(txResponse *rpcClient.TxResponse, useCompressed bool) (*EventResult, error) {
	events := &txResponse.Result.TxResult.ResponseBase.Events
	evCount := len(*events)

	/*
		The reason why the program only compresses events with more than 2 elements is because of the size.
		Even with the trained zstd dictionary in theory we could compress the data but unless it is more than 70 bytes
		altogether in it's raw form then it will just add compression overhead.

		Until the dictionary is improved and trained on a larger set of data for now it will work like this.
	*/
	if useCompressed && evCount >= 2 {
		protoSerializedEv, err := serializeEvent(events)
		if err != nil {
			return nil, err
		}
		compressed := zstdWriter.EncodeAll(protoSerializedEv, nil)
		return &EventResult{
			Format:         CompressedFormat,
			CompressedData: compressed,
		}, nil
	}

	// Native format implementation
	nativeEvents := make([]s.Event, 0, len(txResponse.Result.TxResult.ResponseBase.Events))
	for _, event := range txResponse.Result.TxResult.ResponseBase.Events {
		attributes := make([]s.Attribute, 0, len(event.Attrs))
		for _, attribute := range event.Attrs {
			attributes = append(attributes, s.Attribute{
				Key:   attribute.Key,
				Value: attribute.Value,
			})
		}
		nativeEvents = append(nativeEvents, s.Event{
			AtType:     event.AtType,
			Type:       event.Type,
			Attributes: attributes,
			PkgPath:    event.PkgPath,
		})
	}

	return &EventResult{
		Format:       NativeFormat,
		NativeEvents: nativeEvents,
	}, nil
}

func serializeEvent(events *[]rpcClient.Event) ([]byte, error) {
	protoTxEvents := &events_proto.TxEvents{
		Events: make([]*events_proto.Event, 0),
	}
	for _, event := range *events {
		protoAttrs := make([]*events_proto.Attribute, 0)
		for _, attribute := range event.Attrs {
			protoAttrs = append(protoAttrs, events_proto.NewAttributeFromString(attribute.Key, attribute.Value))
		}
		pkgPath := event.PkgPath
		protoEv := &events_proto.Event{
			AtType:     event.AtType,
			Type:       event.Type,
			Attributes: protoAttrs,
			PkgPath:    &pkgPath,
		}
		protoTxEvents.Events = append(protoTxEvents.Events, protoEv)
	}
	bs, err := proto.Marshal(protoTxEvents)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
