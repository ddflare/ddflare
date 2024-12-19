/*
Copyright © 2024 Francesco Giudici <dev@foggy.day>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ddman

type DNSManager interface {
	Add(fqdn string) error
	Del(fqdn string) error
	GetApiEndpoint() string
	Init(auth string) error
	Resolve(fqdn string) (string, error)
	SetApiEndpoint(ep string)
	Update(fqdn, ip string) error
}
