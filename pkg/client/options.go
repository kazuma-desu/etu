package client

import (
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func buildClientOptions(opts *GetOptions) ([]clientv3.OpOption, error) {
	if opts == nil {
		return nil, nil
	}

	var clientOpts []clientv3.OpOption

	if opts.Prefix {
		clientOpts = append(clientOpts, clientv3.WithPrefix())
	}

	if opts.FromKey {
		clientOpts = append(clientOpts, clientv3.WithFromKey())
	}

	if opts.RangeEnd != "" {
		clientOpts = append(clientOpts, clientv3.WithRange(opts.RangeEnd))
	}

	if opts.Limit > 0 {
		clientOpts = append(clientOpts, clientv3.WithLimit(opts.Limit))
	}

	if opts.Revision > 0 {
		clientOpts = append(clientOpts, clientv3.WithRev(opts.Revision))
	}

	if opts.SortOrder != "" || opts.SortTarget != "" {
		var order clientv3.SortOrder
		var target clientv3.SortTarget

		switch opts.SortOrder {
		case "ASCEND", "":
			order = clientv3.SortAscend
		case "DESCEND":
			order = clientv3.SortDescend
		default:
			return nil, fmt.Errorf("invalid sort order: %s (use ASCEND or DESCEND)", opts.SortOrder)
		}

		switch opts.SortTarget {
		case "KEY", "":
			target = clientv3.SortByKey
		case "VERSION":
			target = clientv3.SortByVersion
		case "CREATE":
			target = clientv3.SortByCreateRevision
		case "MODIFY":
			target = clientv3.SortByModRevision
		case "VALUE":
			target = clientv3.SortByValue
		default:
			return nil, fmt.Errorf("invalid sort target: %s", opts.SortTarget)
		}

		clientOpts = append(clientOpts, clientv3.WithSort(target, order))
	}

	if opts.KeysOnly {
		clientOpts = append(clientOpts, clientv3.WithKeysOnly())
	}

	if opts.CountOnly {
		clientOpts = append(clientOpts, clientv3.WithCountOnly())
	}

	if opts.MinModRev > 0 {
		clientOpts = append(clientOpts, clientv3.WithMinModRev(opts.MinModRev))
	}
	if opts.MaxModRev > 0 {
		clientOpts = append(clientOpts, clientv3.WithMaxModRev(opts.MaxModRev))
	}
	if opts.MinCreateRev > 0 {
		clientOpts = append(clientOpts, clientv3.WithMinCreateRev(opts.MinCreateRev))
	}
	if opts.MaxCreateRev > 0 {
		clientOpts = append(clientOpts, clientv3.WithMaxCreateRev(opts.MaxCreateRev))
	}

	return clientOpts, nil
}
