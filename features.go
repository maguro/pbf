// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf

// The required features of a PBF header that pbf supports. The encoder
// writes these features, and the decoder accepts only these features.
const (
	featureOSMSchema             = "OsmSchema-V0.6"
	featureDenseNodes            = "DenseNodes"
	featureHistoricalInformation = "HistoricalInformation"
)
