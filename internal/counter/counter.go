package counter

type Result struct {
	Count  int
	Labels map[string]string
	Key    string
}

type Counter struct {
	labels []string

	results []*Result
	indexes map[string]int
}

func New(labels []string) *Counter {
	return &Counter{
		labels: labels,

		results: nil,
		indexes: make(map[string]int),
	}
}

func (c *Counter) Increment(labels map[string]string) {
	index := c.ensureIndex(labels)
	c.results[index].Count++
}

func (c *Counter) Set(labels map[string]string, value int) {
	index := c.ensureIndex(labels)
	c.results[index].Count = value
}

func (c *Counter) Results() []*Result {
	return c.results
}

func (c *Counter) GetKey(labels map[string]string) string {
	key := ""
	for _, labelName := range c.labels {
		labelValue, ok := labels[labelName]
		if !ok {
			labelValue = "unknown"
		}

		key += "{" + labelName + ":" + labelValue + "}"
	}

	return key
}

func (c *Counter) ensureIndex(labels map[string]string) int {
	key := c.GetKey(labels)

	index := -1
	if i, ok := c.indexes[key]; ok {
		index = i
	} else {
		c.results = append(c.results, &Result{
			Count:  0,
			Labels: labels,
			Key:    key,
		})
		index = len(c.results) - 1
		c.indexes[key] = index
	}

	return index
}
