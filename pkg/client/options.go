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
		order, target, err := resolveSortOptions(opts.SortOrder, opts.SortTarget)
		if err != nil {
			return nil, err
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

// resolveSortOptions converts string sort order and target to etcd client types.
func resolveSortOptions(sortOrder, sortTarget string) (clientv3.SortOrder, clientv3.SortTarget, error) {
	order, err := parseSortOrder(sortOrder)
	if err != nil {
		return 0, 0, err
	}

	target, err := parseSortTarget(sortTarget)
	if err != nil {
		return 0, 0, err
	}

	return order, target, nil
}

// parseSortOrder converts a sort order string to etcd SortOrder.
func parseSortOrder(order string) (clientv3.SortOrder, error) {
	switch order {
	case "ASCEND", "":
		return clientv3.SortAscend, nil
	case "DESCEND":
		return clientv3.SortDescend, nil
	default:
		return 0, fmt.Errorf("invalid sort order: %s (use ASCEND or DESCEND)", order)
	}
}

// parseSortTarget converts a sort target string to etcd SortTarget.
func parseSortTarget(target string) (clientv3.SortTarget, error) {
	switch target {
	case "KEY", "":
		return clientv3.SortByKey, nil
	case "VERSION":
		return clientv3.SortByVersion, nil
	case "CREATE":
		return clientv3.SortByCreateRevision, nil
	case "MODIFY":
		return clientv3.SortByModRevision, nil
	case "VALUE":
		return clientv3.SortByValue, nil
	default:
		return 0, fmt.Errorf("invalid sort target: %s", target)
	}
}
