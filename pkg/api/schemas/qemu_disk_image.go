package schemas

type QemuDiskImageCreate struct {
	Format        *string `json:"format,omitempty"`
	Size          *int    `json:"size,omitempty"`
	Preallocation *string `json:"preallocation,omitempty"`
	ClusterSize   *int    `json:"cluster_size,omitempty"`
	RefcountBits  *int    `json:"refcount_bits,omitempty"`
	LazyRefcounts *string `json:"lazy_refcounts,omitempty"`
	Subformat     *string `json:"subformat,omitempty"`
	Static        *string `json:"static,omitempty"`
	ZeroedGrain   *string `json:"zeroed_grain,omitempty"`
	AdapterType   *string `json:"adapter_type,omitempty"`
}
