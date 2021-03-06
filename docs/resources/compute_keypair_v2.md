---
subcategory: "Elastic Cloud Server (ECS)"
---

# opentelekomcloud_compute_keypair_v2

Manages a V2 keypair resource within OpenTelekomCloud.

## Example Usage

```hcl
resource "opentelekomcloud_compute_keypair_v2" "test-keypair" {
  name       = "my-keypair"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDAjpC1hwiOCCmKEWxJ4qzTTsJbKzndLotBCz5PcwtUnflmU+gHJtWMZKpuEGVi29h0A/+ydKek1O18k10Ff+4tyFjiHDQAnOfgWf7+b1yK+qDip3X1C0UPMbwHlTfSGWLGZqd9LvEFx9k3h/M+VtMvwR1lJ9LUyTAImnNjWG7TaIPmui30HvM2UiFEmqkr4ijq45MyX2+fLIePLRIF61p4whjHAQYufqyno3BS48icQb4p6iVEZPo4AE2o9oIyQvj2mx4dk5Y8CgSETOZTYDOR3rU2fZTRDRgPJDH9FWvQjF5tA0p3d9CoWWd2s6GKKbfoUIi8R/Db1BSPJwkqB"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the keypair. Changing this creates a new keypair.

* `public_key` - (Required) A pre-generated OpenSSH-formatted public key.
  Changing this creates a new keypair.

->
If both `name` and `public_key` duplicate the existing keypair value, the new keypair won't be
managed by the Terraform. Keypair resource will be marked as `shared.`

* `value_specs` - (Optional) Map of additional options.

## Attributes Reference

The following attributes are exported:

* `name` - See Argument Reference above.

* `public_key` - See Argument Reference above.

* `value_specs` - See Argument Reference above.

* `shared` - Indicates that keypair is shared (global) and not managed by Terraform.

## Import

Keypairs can be imported using the `name`, e.g.

```sh
terraform import opentelekomcloud_compute_keypair_v2.my-keypair test-keypair
```

Imported key pairs are considered to be not shared.
