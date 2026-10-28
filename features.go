// Copyright 2026 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pbf

// The required features of a PBF header that pbf supports. The encoder
// writes these features, and the decoder accepts only these features.
const (
	featureOSMSchema             = "OsmSchema-V0.6"
	featureDenseNodes            = "DenseNodes"
	featureHistoricalInformation = "HistoricalInformation"
)
