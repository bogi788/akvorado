<!-- SPDX-FileCopyrightText: 2022 Free Mobile -->
<!-- SPDX-License-Identifier: AGPL-3.0-only -->

<!--
 Default, 4 columns:
 - Dimensions list box (4 columns)
 - IPv4 /x (optional), IPv6 /x (optional), Limit, Top by
 Large (lg:), 2 columns:
 - Dimensions list box (2 columns)
 - IPv4 /x (optional), IPv6 /x (optional)
 - Limit, Top By
-->

<template>
  <div class="grid grid-cols-4 gap-2 lg:grid-cols-2">
    <InputListBox
      v-model="selectedDimensions"
      :items="dimensions"
      :error="dimensionsError"
      multiple
      label="Dimensions"
      filter="name"
      class="col-span-full"
    >
      <template #selected>
        <span v-if="selectedDimensions.length === 0">No dimensions</span>
        <draggable
          v-model="selectedDimensions"
          class="block flex flex-wrap gap-1"
          tag="span"
          item-key="id"
        >
          <template #item="{ element: dimension }">
            <span
              class="flex cursor-grab items-center gap-1 rounded border-2 bg-violet-100 px-1.5 dark:bg-slate-800 dark:text-gray-200"
              :style="{ borderColor: dimension.color }"
            >
              <span class="leading-4">{{ dimension.name }}</span>
              <XIcon
                class="h-4 w-4 cursor-pointer hover:text-blue-700 dark:hover:text-white"
                @click.stop.prevent="removeDimension(dimension)"
              />
            </span>
          </template>
        </draggable>
        <span
          class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
        >
          <SelectorIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
        </span>
      </template>
      <template #item="{ name, color }">
        <span :style="{ backgroundColor: color }" class="inline w-1 rounded"
          >&nbsp;</span
        >
        {{ name }}
      </template>
    </InputListBox>
    <InputString
      v-if="canAggregate"
      v-model="truncate4"
      class="grow"
      label="IPv4 /x"
      :error="truncate4Error"
    />
    <InputString
      v-if="canAggregate"
      v-model="truncate6"
      class="grow"
      label="IPv6 /x"
      :error="truncate6Error"
    />
    <InputString
      v-model="limit"
      class="grow"
      label="Limit"
      :error="limitError"
    />
    <InputListBox
      v-model="limitType"
      :items="computationModeList"
      class="grow"
      label="Top by"
    >
      <template #selected>{{ limitType.name }}</template>
      <template #item="{ name }">
        <div class="flex w-full items-center justify-between">
          <span>{{ name }}</span>
        </div>
      </template>
    </InputListBox>
  </div>
</template>

<script lang="ts" setup>
import { ref, watch, computed, inject } from "vue";
import draggable from "vuedraggable";
import { XIcon, SelectorIcon } from "@heroicons/vue/solid";
import { dataColor } from "@/utils";
import InputString from "@/components/InputString.vue";
import InputListBox from "@/components/InputListBox.vue";
import { ServerConfigKey } from "@/components/ServerConfigProvider.vue";
import { isEqual, intersection } from "lodash-es";

const props = withDefaults(
  defineProps<{
    modelValue: ModelType;
    minDimensions?: number;
  }>(),
  {
    minDimensions: 0,
  },
);
const emit = defineEmits<{
  "update:modelValue": [value: typeof props.modelValue];
}>();

const serverConfiguration = inject(ServerConfigKey)!;
const selectedDimensions = ref<Array<(typeof dimensions.value)[0]>>([]);
const dimensionsError = computed(() => {
  if (selectedDimensions.value.length < props.minDimensions) {
    return "At least two dimensions are required";
  }
  return "";
});
const limit = ref("10");
const limitError = computed(() => {
  const val = parseInt(limit.value);
  if (isNaN(val)) {
    return "Not a number";
  }
  if (val < 1) {
    return "Should be ≥ 1";
  }
  const upperLimit = serverConfiguration.value?.dimensionsLimit ?? 50;
  if (val > upperLimit) {
    return `Should be ≤ ${upperLimit}`;
  }
  return "";
});

const computationModes = {
  avg: "Avg",
  max: "Max",
  last: "Last",
} as const;

const computationModeList = Object.entries(computationModes).map(
  ([k, v], idx) => ({
    id: idx + 1,
    type: k as keyof typeof computationModes, // why isn't it infered?
    name: v,
  }),
);
const limitType = ref(computationModeList[0]);

const canAggregate = computed(
  () =>
    intersection(
      selectedDimensions.value.map((dim) => dim.name),
      serverConfiguration.value?.truncatable || [],
    ).length > 0,
);
const truncate4 = ref("32");
const truncate4Error = computed(() => {
  const val = parseInt(truncate4.value);
  if (isNaN(val) || val <= 0 || val > 32) {
    return "0 < x ≤ 32";
  }
  return "";
});
const truncate6 = ref("128");
const truncate6Error = computed(() => {
  const val = parseInt(truncate6.value);
  if (isNaN(val) || val <= 0 || val > 128) {
    return "0 < x ≤ 128";
  }
  return "";
});
const hasErrors = computed(
  () =>
    !!limitError.value ||
    !!dimensionsError.value ||
    !!truncate4Error.value ||
    !!truncate6Error.value,
);

const dimensions = computed(
  () =>
    serverConfiguration.value?.dimensions.map((v, idx) => ({
      id: idx + 1,
      name: v,
      color: dataColor(
        ["Exporter", "Src", "Dst", "In", "Out", ""]
          .map((p) => v.startsWith(p))
          .indexOf(true),
      ),
    })) || [],
);

const removeDimension = (dimension: (typeof dimensions.value)[0]) => {
  selectedDimensions.value = selectedDimensions.value.filter(
    (d) => d !== dimension,
  );
};
watch(
  () => [props.modelValue, dimensions.value] as const,
  ([value, dimensions]) => {
    if (value) {
      limit.value = value.limit.toString();
      limitType.value =
        computationModeList.find((mode) => mode.type === value.limitType) ||
        computationModeList[0];
      truncate4.value = value.truncate4.toString();
      truncate6.value = value.truncate6.toString();
    }
    if (value)
      selectedDimensions.value = value.selected
        .map((name) => dimensions.find((d) => d.name === name))
        .filter((d): d is (typeof dimensions)[0] => !!d);
  },
  { immediate: true, deep: true },
);
watch(
  [
    selectedDimensions,
    limit,
    limitType,
    truncate4,
    truncate6,
    hasErrors,
  ] as const,
  ([selected, limit, limitType, truncate4, truncate6, hasErrors]) => {
    const updated = {
      selected: selected.map((d) => d.name),
      limit: parseInt(limit),
      limitType: limitType.type,
      truncate4: parseInt(truncate4),
      truncate6: parseInt(truncate6),
      errors: hasErrors,
    };
    if (
      !isEqual(updated, props.modelValue) &&
      !isNaN(updated.limit) &&
      !isNaN(updated.truncate4) &&
      !isNaN(updated.truncate6)
    ) {
      emit("update:modelValue", updated);
    }
  },
);
</script>

<script lang="ts">
export type ModelType = {
  selected: string[];
  limit: number;
  limitType: string;
  truncate4: number;
  truncate6: number;
  errors?: boolean;
} | null;
</script>
