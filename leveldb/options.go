// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func dupOptions(o *opt.Options) *opt.Options {
	newo := &opt.Options{}
	if o != nil {
		*newo = *o
	}
	if newo.Strict == 0 {
		newo.Strict = opt.DefaultStrict
	}
	return newo
}

func (s *session) setOptions(o *opt.Options) {
	no := dupOptions(o)
/**
	no: &opt.Options{
		Filter:      filter.NewBloomFilter(10),
		DisableSeeksCompaction: true,
		OpenFilesCacheCapacity: 16
		BlockCacheCapacity: 8MiB
		WriteBuffer: 4MiB,
		Strict: StrictJournalChecksum | 
			StrictBlockChecksum | 
			StrictCompaction | 
			StrictReader,
	}
*/
	// Alternative filters.
	if filters := o.GetAltFilters(); len(filters) > 0 {
		no.AltFilters = make([]filter.Filter, len(filters))
		for i, filter := range filters {
			no.AltFilters[i] = &iFilter{filter}
		}
	}
	// Comparer.
	s.icmp = &iComparer{o.GetComparer()}
	no.Comparer = s.icmp
	// Filter.
	if filter := o.GetFilter(); filter != nil {
		no.Filter = &iFilter{filter}
	}

/**
	no: &opt.Options{
		Filter:      filter.NewBloomFilter(10),
		DisableSeeksCompaction: true,
		OpenFilesCacheCapacity: 16
		BlockCacheCapacity: 8MiB
		WriteBuffer: 4MiB,
		Strict: StrictJournalChecksum | 
			StrictBlockChecksum | 
			StrictCompaction | 
			StrictReader,
		Comparer: &iComparer{ // 同s指向
			ucmp: bytesComparer{}
		},
		Filter: &iFilter{
			Filter: 就是外层的BoolmFilter(10)
		},
	}
*/
	s.o = &cachedOptions{Options: no}
	s.o.cache()
/**
	cache完
	o: &cachedOptions{
		Options: no,
		compactionExpandLimit []int:
		[ 2MiB*25, 剩余6个一样的 ]

		compactionGPOverlaps  []int:
		[ 2MiB*10, 剩余6个一样的 ]

		compactionSourceLimit []int:
		[ 2MiB*1, 剩余6个一样的 ]

		compactionTableSize   []int:
		[ 2MiB*1, 剩余6个一样的 ]

		compactionTotalSize   []int64:
		[ 10MiB*10^0, 10MiB*10^1, ..., 10MiB*10^6]
	}
*/
}

const optCachedLevel = 7

type cachedOptions struct {
	*opt.Options

	compactionExpandLimit []int
	compactionGPOverlaps  []int
	compactionSourceLimit []int
	compactionTableSize   []int
	compactionTotalSize   []int64
}

func (co *cachedOptions) cache() {
	co.compactionExpandLimit = make([]int, optCachedLevel)
	co.compactionGPOverlaps = make([]int, optCachedLevel)
	co.compactionSourceLimit = make([]int, optCachedLevel)
	co.compactionTableSize = make([]int, optCachedLevel)
	co.compactionTotalSize = make([]int64, optCachedLevel)

	for level := 0; level < optCachedLevel; level++ {
		co.compactionExpandLimit[level] = co.Options.GetCompactionExpandLimit(level)
		co.compactionGPOverlaps[level] = co.Options.GetCompactionGPOverlaps(level)
		co.compactionSourceLimit[level] = co.Options.GetCompactionSourceLimit(level)
		co.compactionTableSize[level] = co.Options.GetCompactionTableSize(level)
		co.compactionTotalSize[level] = co.Options.GetCompactionTotalSize(level)
	}
}

func (co *cachedOptions) GetCompactionExpandLimit(level int) int {
	if level < optCachedLevel {
		return co.compactionExpandLimit[level]
	}
	return co.Options.GetCompactionExpandLimit(level)
}

func (co *cachedOptions) GetCompactionGPOverlaps(level int) int {
	if level < optCachedLevel {
		return co.compactionGPOverlaps[level]
	}
	return co.Options.GetCompactionGPOverlaps(level)
}

func (co *cachedOptions) GetCompactionSourceLimit(level int) int {
	if level < optCachedLevel {
		return co.compactionSourceLimit[level]
	}
	return co.Options.GetCompactionSourceLimit(level)
}

func (co *cachedOptions) GetCompactionTableSize(level int) int {
	if level < optCachedLevel {
		return co.compactionTableSize[level]
	}
	return co.Options.GetCompactionTableSize(level)
}

func (co *cachedOptions) GetCompactionTotalSize(level int) int64 {
	if level < optCachedLevel {
		// [ 10MiB*10^0, 10MiB*10^1, ..., 10MiB*10^6 ]
		return co.compactionTotalSize[level]
	}
	return co.Options.GetCompactionTotalSize(level)
}
