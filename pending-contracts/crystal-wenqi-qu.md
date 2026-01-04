# Contract: Wenqi Qu Qu (Crystal)

**Status:** Client Signed - Awaiting contractor signature via Deel web UI
**Contract ID:** `mv9g5k8`
**Created:** 2026-01-03

## Worker Details

| Field | Value |
|-------|-------|
| Name | Wenqi Qu Qu (Crystal) |
| Email | quwenqi51@outlook.com |
| Phone | +16042091231 |
| Country | Canada (CA) |
| Currency | CAD |
| Role | Host / Livestreamer |
| Notion Page | [Guide to Crystal](https://www.notion.so/Guide-to-Crystal-2dceaeccf76481eeabe7c260d9bb7225) |
| Notion Worker ID | 294eaecc-f764-8177-b125-ceb738a64008 |

## Contract Details

| Field | Value |
|-------|-------|
| Contract ID | `mv9g5k8` |
| Contract Type | Pay As You Go (`pay_as_you_go_time_based`) |
| Template | 🇨🇦 Host (`1034f737-9ead-41d3-8137-d47391c16951`) |
| Legal Entity | Delicious Milk Corporation (`894ea227-4d2c-4d35-87a5-8fddc5b0aef7`) |
| Group | North America (`f9f466f3-13ef-4ff9-85a2-5d01a6637627`) |
| Start Date | 2026-01-05 (Monday) |
| End Date | None |
| Payment Cycle | 5th of month (monthly) |
| Base Rate | $23.00 CAD/hour |
| Status | `waiting_for_contractor_sign` |

## Compensation Structure

**Base Hourly Rate:** $23.00 CAD per hour during all live streaming periods

**GMV Commission Structure (Cumulative Progressive Tiers):**
- Tier 1: $0-$4,999 GMV → 0.25% commission
- Tier 2: $5,000-$14,999 GMV → 0.30% commission
- Tier 3: $15,000-$29,999 GMV → 0.40% commission
- Tier 4: $30,000-$49,999 GMV → 0.50% commission
- Tier 5: $50,000-$69,999 GMV → 1.00% commission
- Tier 6: $70,000-$89,999 GMV → 1.50% commission
- Tier 7: $90,000-$129,999 GMV → 2.00% commission
- Tier 8: $130,000-$179,999 GMV → 2.50% commission
- Tier 9: $180,000+ GMV → 3.00% commission

## Scope of Work

Full scope available in Notion at the worker's "Role Scope" property. Includes:
- Stream scheduling (min 3/week)
- Product knowledge and preparation
- Live selling performance (min $25,000 GMV per 3-hour stream)
- Audience engagement
- Inventory and sample management
- Vendor relations
- Team coordination
- Content creation
- Customer service
- Performance tracking
- Cross-border operations (if applicable)
- Training and mentorship

## Manager Details

| Field | Value |
|-------|-------|
| Manager | Vera |
| Manager ID | `fd470477-d950-47dd-93eb-d31830d6caca` |
| Relationship | Direct Report |

**Note:** Manager will be set via worker-relations API after Crystal accepts the contract and has a profile ID.

## CLI Command Used

```bash
./deel contracts create --account wanver \
  --title "Livestreamer Contractor Agreement - Wenqi Qu Qu (Crystal)" \
  --type pay_as_you_go_time_based \
  --worker-email "quwenqi51@outlook.com" \
  --worker-first "Wenqi" \
  --worker-last "Qu Qu" \
  --currency CAD \
  --country CA \
  --rate 23.00 \
  --job-title "Host / Livestreamer" \
  --start-date "2026-01-05" \
  --template "1034f737-9ead-41d3-8137-d47391c16951" \
  --legal-entity "894ea227-4d2c-4d35-87a5-8fddc5b0aef7" \
  --group "f9f466f3-13ef-4ff9-85a2-5d01a6637627" \
  --cycle-end 5 \
  --cycle-end-type DAY_OF_MONTH \
  --frequency monthly \
  --special-clause "GMV Commission Structure..."

# Sign contract
./deel contracts sign mv9g5k8 --account wanver --signer "Vladimir Novosselov"
```

## Next Steps

1. ~~Create contract via CLI~~ Done - Contract ID: `mv9g5k8`
2. ~~Sign contract (client side)~~ Done - via CLI with `--signer` flag
3. **Send invitation to Crystal** - Must be done via Deel web UI (API returns 404)
4. **Set manager to Vera** - After Crystal accepts, run:
   ```bash
   ./deel worker-relations create \
     --profile-id <CRYSTAL_PROFILE_ID> \
     --manager-id fd470477-d950-47dd-93eb-d31830d6caca \
     --relation-type direct_report \
     --start-date 2026-01-05
   ```

## Notes

- The Deel API invite endpoints return 404 for pay-as-you-go contract types
- Invitation must be completed through the Deel web interface
- Contract is visible at: https://app.deel.com/contract/mv9g5k8/contracts
- Manager relationship requires profile ID (available after worker accepts)
