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

	// TODO
}

func NewWasteManager(isRecyclable RecyclingCriterion, recycler Recycler) WasteManager {
	// TODO
	return nil
}
