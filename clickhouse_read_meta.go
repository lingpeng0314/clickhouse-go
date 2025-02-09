package clickhouse

import (
	"fmt"

	"github.com/ClickHouse/clickhouse-go/lib/data"
	"github.com/ClickHouse/clickhouse-go/lib/protocol"
)

func (ch *clickhouse) readMeta() (*data.Block, error) {
	for {
		packet, err := ch.decoder.Uvarint()
		if err != nil {
			return nil, err
		}

		switch packet {
		case protocol.ServerException:
			ch.logf("[read meta] <- exception")
			return nil, ch.exception()
		case protocol.ServerProgress:
			progress := &Progress{}
			err := progress.update(ch)
			if err != nil {
				return nil, err
			}
			ch.logf("[read meta] <- progress: rows=%d, bytes=%d, total rows=%d",
				progress.Rows,
				progress.Bytes,
				progress.TotalRows,
			)
		case protocol.ServerProfileInfo:
			profileInfo, err := ch.profileInfo()
			if err != nil {
				return nil, err
			}
			ch.logf("[read meta] <- profiling: rows=%d, bytes=%d, blocks=%d", profileInfo.rows, profileInfo.bytes, profileInfo.blocks)
		case protocol.ServerData:
			block, err := ch.readBlock()
			if err != nil {
				return nil, err
			}
			ch.logf("[read meta] <- data: packet=%d, columns=%d, rows=%d", packet, block.NumColumns, block.NumRows)
			return block, nil
		case protocol.ServerEndOfStream:
			_, err := ch.readBlock()
			ch.logf("[process] <- end of stream")
			return nil, err
		default:
			ch.conn.Close()
			return nil, fmt.Errorf("[read meta] unexpected packet [%d] from server", packet)
		}
	}
}
