package monitor

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/monitor/mgmt/2020-10-01/insights"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceMonitorActivityLogAlert() *schema.Resource {
	return &schema.Resource{
		Create: resourceMonitorActivityLogAlertCreateUpdate,
		Read:   resourceMonitorActivityLogAlertRead,
		Update: resourceMonitorActivityLogAlertCreateUpdate,
		Delete: resourceMonitorActivityLogAlertDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"scopes": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				Set: schema.HashString,
			},

			"criteria": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"category": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Administrative",
								"Autoscale",
								"Policy",
								"Recommendation",
								"ResourceHealth",
								"Security",
								"ServiceHealth",
							}, false),
						},
						"operation_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"caller": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"level": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Verbose",
								"Informational",
								"Warning",
								"Error",
								"Critical",
							}, false),
						},
						"resource_provider": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resource_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resource_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resource_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"status": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"sub_status": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"recommendation_category": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Cost",
								"Reliability",
								"OperationalExcellence",
								"Performance",
							},
								false,
							),
							ConflictsWith: []string{"criteria.0.recommendation_type"},
						},
						"recommendation_impact": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"High",
								"Medium",
								"Low",
							},
								false,
							),
							ConflictsWith: []string{"criteria.0.recommendation_type"},
						},
						"recommendation_type": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"criteria.0.recommendation_category", "criteria.0.recommendation_impact"},
						},
					},
				},
			},

			"action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_group_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"webhook_properties": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
				Set: resourceMonitorActivityLogAlertActionHash,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceMonitorActivityLogAlertCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActivityLogAlertsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Monitor Activity Log Alert %q (Resource Group %q): %s", name, resourceGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_monitor_activity_log_alert", *existing.ID)
		}
	}

	enabled := d.Get("enabled").(bool)
	description := d.Get("description").(string)
	scopesRaw := d.Get("scopes").(*schema.Set).List()
	criteriaRaw := d.Get("criteria").([]interface{})
	actionRaw := d.Get("action").(*schema.Set).List()

	t := d.Get("tags").(map[string]interface{})
	expandedTags := tags.Expand(t)

	parameters := insights.ActivityLogAlertResource{
		Location: utils.String(azure.NormalizeLocation("Global")),
		AlertRuleProperties: &insights.AlertRuleProperties{
			Enabled:     utils.Bool(enabled),
			Description: utils.String(description),
			Scopes:      utils.ExpandStringSlice(scopesRaw),
			Condition:   expandMonitorActivityLogAlertCriteria(criteriaRaw),
			Actions:     expandMonitorActivityLogAlertAction(actionRaw),
		},
		Tags: expandedTags,
	}

	if _, err := client.CreateOrUpdate(ctx, resourceGroup, name, parameters); err != nil {
		return fmt.Errorf("Error creating or updating activity log alert %q (resource group %q): %+v", name, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Activity log alert %q (resource group %q) ID is empty", name, resourceGroup)
	}
	d.SetId(*read.ID)

	return resourceMonitorActivityLogAlertRead(d, meta)
}

func resourceMonitorActivityLogAlertRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActivityLogAlertsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	name := id.Path["activityLogAlerts"]

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Activity Log Alert %q was not found in Resource Group %q - removing from state!", name, resourceGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error getting activity log alert %q (resource group %q): %+v", name, resourceGroup, err)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resourceGroup)
	if alert := resp.AlertRuleProperties; alert != nil {
		d.Set("enabled", alert.Enabled)
		d.Set("description", alert.Description)
		if err := d.Set("scopes", utils.FlattenStringSlice(alert.Scopes)); err != nil {
			return fmt.Errorf("Error setting `scopes`: %+v", err)
		}
		if err := d.Set("criteria", flattenMonitorActivityLogAlertCriteria(alert.Condition)); err != nil {
			return fmt.Errorf("Error setting `criteria`: %+v", err)
		}
		if err := d.Set("action", flattenMonitorActivityLogAlertAction(alert.Actions)); err != nil {
			return fmt.Errorf("Error setting `action`: %+v", err)
		}
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceMonitorActivityLogAlertDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Monitor.ActivityLogAlertsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	name := id.Path["activityLogAlerts"]

	if resp, err := client.Delete(ctx, resourceGroup, name); err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("Error deleting activity log alert %q (resource group %q): %+v", name, resourceGroup, err)
		}
	}

	return nil
}

func expandMonitorActivityLogAlertCriteria(input []interface{}) *insights.AlertRuleAllOfCondition {
	conditions := make([]insights.AlertRuleAnyOfOrLeafCondition, 0)
	v := input[0].(map[string]interface{})

	if category := v["category"].(string); category != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("category"),
			Equals: utils.String(category),
		})
	}
	if op := v["operation_name"].(string); op != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("operationName"),
			Equals: utils.String(op),
		})
	}
	if caller := v["caller"].(string); caller != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("caller"),
			Equals: utils.String(caller),
		})
	}
	if level := v["level"].(string); level != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("level"),
			Equals: utils.String(level),
		})
	}
	if resourceProvider := v["resource_provider"].(string); resourceProvider != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("resourceProvider"),
			Equals: utils.String(resourceProvider),
		})
	}
	if resourceType := v["resource_type"].(string); resourceType != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("resourceType"),
			Equals: utils.String(resourceType),
		})
	}
	if resourceGroup := v["resource_group"].(string); resourceGroup != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("resourceGroup"),
			Equals: utils.String(resourceGroup),
		})
	}
	if id := v["resource_id"].(string); id != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("resourceId"),
			Equals: utils.String(id),
		})
	}
	if status := v["status"].(string); status != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("status"),
			Equals: utils.String(status),
		})
	}
	if subStatus := v["sub_status"].(string); subStatus != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("subStatus"),
			Equals: utils.String(subStatus),
		})
	}
	if recommendationType := v["recommendation_type"].(string); recommendationType != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("properties.recommendationType"),
			Equals: utils.String(recommendationType),
		})
	}

	if recommendationCategory := v["recommendation_category"].(string); recommendationCategory != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("properties.recommendationCategory"),
			Equals: utils.String(recommendationCategory),
		})
	}

	if recommendationImpact := v["recommendation_impact"].(string); recommendationImpact != "" {
		conditions = append(conditions, insights.AlertRuleAnyOfOrLeafCondition{
			Field:  utils.String("properties.recommendationImpact"),
			Equals: utils.String(recommendationImpact),
		})
	}

	return &insights.AlertRuleAllOfCondition{
		AllOf: &conditions,
	}
}

func expandMonitorActivityLogAlertAction(input []interface{}) *insights.ActionList {
	actions := make([]insights.ActionGroup, 0)
	for _, item := range input {
		v := item.(map[string]interface{})
		if agID := v["action_group_id"].(string); agID != "" {
			props := make(map[string]*string)
			if pVal, ok := v["webhook_properties"]; ok {
				for pk, pv := range pVal.(map[string]interface{}) {
					props[pk] = utils.String(pv.(string))
				}
			}

			actions = append(actions, insights.ActionGroup{
				ActionGroupID:     utils.String(agID),
				WebhookProperties: props,
			})
		}
	}
	return &insights.ActionList{
		ActionGroups: &actions,
	}
}

func flattenMonitorActivityLogAlertCriteria(input *insights.ActivityLogAlertAllOfCondition) []interface{} {
	result := make(map[string]interface{})
	if input == nil || input.AllOf == nil {
		return []interface{}{result}
	}
	for _, condition := range *input.AllOf {
		if condition.Field != nil && condition.Equals != nil {
			switch strings.ToLower(*condition.Field) {
			case "operationname":
				result["operation_name"] = *condition.Equals
			case "resourceprovider":
				result["resource_provider"] = *condition.Equals
			case "resourcetype":
				result["resource_type"] = *condition.Equals
			case "resourcegroup":
				result["resource_group"] = *condition.Equals
			case "resourceid":
				result["resource_id"] = *condition.Equals
			case "substatus":
				result["sub_status"] = *condition.Equals
			case "properties.recommendationtype":
				result["recommendation_type"] = *condition.Equals
			case "properties.recommendationcategory":
				result["recommendation_category"] = *condition.Equals
			case "properties.recommendationimpact":
				result["recommendation_impact"] = *condition.Equals
			case "caller", "category", "level", "status":
				result[*condition.Field] = *condition.Equals
			}
		}
	}
	return []interface{}{result}
}

func flattenMonitorActivityLogAlertAction(input *insights.ActivityLogAlertActionList) (result []interface{}) {
	result = make([]interface{}, 0)
	if input == nil || input.ActionGroups == nil {
		return
	}
	for _, action := range *input.ActionGroups {
		v := make(map[string]interface{})

		if action.ActionGroupID != nil {
			v["action_group_id"] = *action.ActionGroupID
		}

		props := make(map[string]string)
		for pk, pv := range action.WebhookProperties {
			if pv != nil {
				props[pk] = *pv
			}
		}
		v["webhook_properties"] = props

		result = append(result, v)
	}
	return result
}

func resourceMonitorActivityLogAlertActionHash(input interface{}) int {
	var buf bytes.Buffer
	if v, ok := input.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%s-", v["action_group_id"].(string)))
	}
	return schema.HashString(buf.String())
}
