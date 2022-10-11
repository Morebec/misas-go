// Copyright 2022 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prediction

import (
	"context"
	"github.com/morebec/misas-go/misas/clock"
	"github.com/pkg/errors"
	"time"
)

// ProcessorOptions Represents the operating options of the Processor.
type ProcessorOptions struct {
	RemovalStrategy ProcessorRemovalStrategy

	// Represents the interval of time to wait between every check to find out new predictions that need to be processed.
	// This value should be tweaked according to the requirements of the system.
	// For example, some systems have predictions that can only happen on a weekly basis, meaning that performing a check every
	// day would likely be sufficient.
	// This value would represent the potential duration delay that it would take between a prediction occurring and for it to
	// be acknowledged/processed by the system.
	SleepDuration time.Duration
}

func NewDefaultProcessorOptions() ProcessorOptions {
	return ProcessorOptions{
		RemovalStrategy: RemoveAfterProcessing,
		SleepDuration:   time.Second * 30,
	}
}

// ProcessorRemovalStrategy Represents the strategy to use for removing Prediction from the store in regard to their processing.
type ProcessorRemovalStrategy string

// RemoveBeforeProcessing Indicates that the Prediction should be removed from the store **before** processing it.
// **Use this for at-most-once delivery.**
const RemoveBeforeProcessing ProcessorRemovalStrategy = "RemoveBeforeProcessing"

// RemoveAfterProcessing Indicates that the Prediction should be removed from the store **after** processing it.
// **Use this for at-least-once delivery.**
const RemoveAfterProcessing ProcessorRemovalStrategy = "RemoveAfterProcessing"

// Processor System responsible for checking if a Prediction has occurred and to send it to the
// InMemoryBus.
type Processor struct {
	clock          clock.Clock
	options        ProcessorOptions
	store          Store
	processingFunc ProcessingFunc
}

// NewProcessor Creates a new Processor.
func NewProcessor(clock clock.Clock, options ProcessorOptions, store Store, processingFunc ProcessingFunc) *Processor {
	if options.RemovalStrategy == "" {
		options.RemovalStrategy = RemoveAfterProcessing
	}

	if processingFunc == nil {
		panic("cannot create a processor without a processing func.")
	}

	if store == nil {
		panic("cannot create a processor without a store.")
	}

	return &Processor{clock: clock, options: options, store: store, processingFunc: processingFunc}
}

func NewSendToBusProcessor(clock clock.Clock, store Store, converter *Converter, options ProcessorOptions, bus *InMemoryBus) *Processor {
	return &Processor{clock: clock, options: options, store: store, processingFunc: SendToPredictionBusProcessingFunc(bus, converter)}
}

type ProcessingFunc func(p Descriptor, ctx context.Context) error

func (p *Processor) Run(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			if err := p.findAndProcessPredictions(ctx); err != nil {
				return errors.Wrap(err, "failed processing prediction")
			}
			time.Sleep(p.options.SleepDuration)
		}

	}
}

func (p *Processor) findAndProcessPredictions(ctx context.Context) error {
	descriptors, err := p.store.FindOccurredBefore(p.clock.Now())
	if err != nil {
		return err
	}

	for _, descriptor := range descriptors {
		if p.options.RemovalStrategy == RemoveBeforeProcessing {
			if err := p.store.Remove(descriptor.ID); err != nil {
				return err
			}
		}

		if err := p.processingFunc(descriptor, ctx); err != nil {
			return err
		}

		if p.options.RemovalStrategy == RemoveAfterProcessing {
			if err := p.store.Remove(descriptor.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// SendToPredictionBusProcessingFunc Builds a Processing Func that sends the Prediction to the InMemoryBus.
func SendToPredictionBusProcessingFunc(bus Bus, converter *Converter) ProcessingFunc {
	return func(descriptor Descriptor, ctx context.Context) error {
		p, err := converter.FromDescriptor(descriptor)
		if err != nil {
			return err
		}
		if err := bus.Send(ctx, p); err != nil {
			return err
		}
		return nil
	}
}
