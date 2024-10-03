package dispatcher

type Scrap string
type Waste string
type RecycledGood string

// Function that determines if a scrap is recyclable. Also returns, if the scrap is not recyclable, the waste it becomes.
type RecyclingCriterion func(Scrap) (isRecyclable bool, asWaste Waste)

// Function that recycles a scrap into a RecycledGood.
type Recycler func(Scrap) RecycledGood

type WasteManager interface {
	ChangeRecyclingCriterion(RecyclingCriterion)
	ChangeRecycler(Recycler)

	Process(Scrap)
	NextRecycledGood() RecycledGood
	NextWaste() Waste
}

type wasteManager struct {
	isRecyclable RecyclingCriterion
	recycler     Recycler

	newRecycler       chan Recycler
	newWasteCriterion chan RecyclingCriterion
	scraps            chan Scrap
	recycledGoods     chan RecycledGood
	waste             chan Waste
}

func NewWasteManager(isRecyclable RecyclingCriterion, recycler Recycler) WasteManager {
	d := &wasteManager{
		isRecyclable:      isRecyclable,
		recycler:          recycler,
		newRecycler:       make(chan Recycler),
		newWasteCriterion: make(chan RecyclingCriterion),
		scraps:            make(chan Scrap),
		recycledGoods:     make(chan RecycledGood),
		waste:             make(chan Waste),
	}
	go d.process()
	return d
}

// Represents the promise of the future delivery of a recycled good.
type RecycledGoodPromise <-chan RecycledGood

// Recycles scraps concurrently, and buffers the recycled goods until they are consumed.
func (m *wasteManager) bufferedRecycle(goodsToBuffer chan RecycledGoodPromise) {
	// Storing the channels themselves, on which the results of recycling will be sent.
	// This keeps the order, while allowing for concurrent recycling, which may not always finish in the same order.
	buffer := make([]<-chan RecycledGood, 0, 100)

	var nextRecycled *RecycledGood = nil

	for {
		if len(buffer) == 0 {
			// If there is nothing buffered, only event is a scrap to recycle.
			goodChan := <-goodsToBuffer
			buffer = append(buffer, goodChan)
		}
		if nextRecycled == nil {
			// Wait for the next recycled good to be ready to be sent.
			// Note that we could add a for loop here too, to keep receiving scraps to buffer while waiting.
			good := <-buffer[0]
			nextRecycled = &good
		}
		select {
		case goodChan := <-goodsToBuffer:
			buffer = append(buffer, goodChan)
		case m.recycledGoods <- *nextRecycled:
			buffer = buffer[1:]
			nextRecycled = nil
		}
	}
}

func (m *wasteManager) bufferWaste(wasteToBuffer <-chan Waste) {
	buffer := make([]Waste, 0, 100)
	for {
		if len(buffer) == 0 {
			waste := <-wasteToBuffer
			buffer = append(buffer, waste)
		}
		select {
		case waste := <-wasteToBuffer:
			buffer = append(buffer, waste)
		case m.waste <- buffer[0]:
			buffer = buffer[1:]
		}
	}
}

func (m *wasteManager) process() {
	goodsToBuffer := make(chan RecycledGoodPromise)
	wasteToBuffer := make(chan Waste)

	go m.bufferedRecycle(goodsToBuffer)
	go m.bufferWaste(wasteToBuffer)

	for {
		select {
		case scrap := <-m.scraps:
			if isRec, w := m.isRecyclable(scrap); !isRec {
				wasteToBuffer <- w
			} else {
				goodChan := make(chan RecycledGood)
				// One goroutine per call to recycler, enabling concurrent execution.
				// Passing recycler by value to avoid concurrent access to m
				go func(recycler Recycler) {
					good := recycler(scrap)
					goodChan <- good
				}(m.recycler)
				goodsToBuffer <- goodChan
			}
		case criterion := <-m.newWasteCriterion:
			m.isRecyclable = criterion
		case recycler := <-m.newRecycler:
			m.recycler = recycler
		}
	}
}

func (m *wasteManager) ChangeRecyclingCriterion(isWaste RecyclingCriterion) {
	m.newWasteCriterion <- isWaste
}

func (m *wasteManager) ChangeRecycler(recycler Recycler) {
	m.newRecycler <- recycler
}

func (m *wasteManager) Process(scrap Scrap) {
	m.scraps <- scrap
}

func (m *wasteManager) NextRecycledGood() RecycledGood {
	return <-m.recycledGoods
}

func (m *wasteManager) NextWaste() Waste {
	return <-m.waste
}
