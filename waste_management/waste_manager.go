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
	recyclables       chan Scrap
	recycledGoods     chan RecycledGood
	waste             chan Waste
}

func NewWasteManager(isRecyclable RecyclingCriterion, recycler Recycler) WasteManager {
	d := &wasteManager{
		isRecyclable:      isRecyclable,
		recycler:          recycler,
		newRecycler:       make(chan Recycler),
		newWasteCriterion: make(chan RecyclingCriterion),
		scraps:            make(chan Scrap, 1000),
		recyclables:       make(chan Scrap),
		recycledGoods:     make(chan RecycledGood),
		waste:             make(chan Waste),
	}
	go d.process()
	go d.recycle()
	return d
}

func (m *wasteManager) recycle() {
	for scrap := range m.recyclables {
		m.recycledGoods <- m.recycler(scrap)
	}
}

func (m *wasteManager) process() {
	for {
		select {
		case scrap := <-m.scraps:
			if isRec, w := m.isRecyclable(scrap); !isRec {
				m.waste <- w
			} else {
				m.recyclables <- scrap
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
