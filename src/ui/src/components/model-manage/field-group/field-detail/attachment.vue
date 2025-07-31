<!--
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
-->

<template>
  <div class="attachment-field-config">
    <div class="form-label cmdb-form-item" :class="{ 'is-error': errors.has('maxFileSize') }">
      <span class="label-text">{{$t('文件大小限制')}}</span>
      <div class="file-size-config">
        <bk-input
          type="number"
          name="maxFileSize"
          v-model.number="localValue.maxFileSize"
          :disabled="isReadOnly"
          :min="1"
          :max="100"
          v-validate="'required|min:1|max:100'"
          @input="handleInput">
        </bk-input>
        <span class="unit">MB</span>
      </div>
      <p class="form-error">{{errors.first('maxFileSize')}}</p>
    </div>

    <div class="form-label cmdb-form-item" :class="{ 'is-error': errors.has('maxFileCount') }">
      <span class="label-text">{{$t('最大文件数量')}}</span>
      <bk-input
        type="number"
        name="maxFileCount"
        v-model.number="localValue.maxFileCount"
        :disabled="isReadOnly"
        :min="1"
        :max="20"
        v-validate="'required|min:1|max:20'"
        @input="handleInput">
      </bk-input>
      <p class="form-error">{{errors.first('maxFileCount')}}</p>
    </div>

    <div class="form-label cmdb-form-item">
      <span class="label-text">{{$t('允许的文件类型')}}</span>
      <div class="file-types-config">
        <bk-checkbox-group v-model="localValue.allowedTypes" @change="handleInput">
          <bk-checkbox value="image" :disabled="isReadOnly">
            {{$t('图片')}} (.jpg, .jpeg, .png, .gif, .bmp, .webp)
          </bk-checkbox>
          <bk-checkbox value="document" :disabled="isReadOnly">
            {{$t('文档')}} (.pdf, .txt)
          </bk-checkbox>
          <bk-checkbox value="archive" :disabled="isReadOnly">
            {{$t('压缩包')}} (.zip)
          </bk-checkbox>
        </bk-checkbox-group>
      </div>
    </div>

    <div class="form-label cmdb-form-item">
      <span class="label-text">{{$t('存储路径')}}</span>
      <bk-input
        type="text"
        v-model="localValue.storagePath"
        :disabled="isReadOnly"
        :placeholder="$t('留空使用默认路径 /data/attachments/')"
        @input="handleInput">
      </bk-input>
    </div>
  </div>
</template>

<script>
  export default {
    props: {
      value: {
        type: Object,
        default: () => ({})
      },
      isReadOnly: {
        type: Boolean,
        default: false
      }
    },
    data() {
      return {
        localValue: {
          maxFileSize: 10, // 默认10MB
          maxFileCount: 5, // 默认最多5个文件
          allowedTypes: ['image', 'document'], // 默认允许图片和文档
          storagePath: '' // 默认使用系统路径
        }
      }
    },
    watch: {
      value: {
        handler(newValue) {
          this.localValue = Object.assign({
            maxFileSize: 10,
            maxFileCount: 5,
            allowedTypes: ['image', 'document'],
            storagePath: ''
          }, newValue)
        },
        immediate: true,
        deep: true
      }
    },
    methods: {
      handleInput() {
        this.$emit('input', this.localValue)
      },
      validate() {
        return this.$validator.validateAll()
      }
    }
  }
</script>

<style lang="scss" scoped>
  .attachment-field-config {
    .file-size-config {
      display: flex;
      align-items: center;
      
      .unit {
        margin-left: 8px;
        color: #63656e;
      }
    }

    .file-types-config {
      margin-top: 8px;
      
      .bk-form-checkbox {
        display: block;
        margin-bottom: 8px;
        
        &:last-child {
          margin-bottom: 0;
        }
      }
    }
  }
</style>