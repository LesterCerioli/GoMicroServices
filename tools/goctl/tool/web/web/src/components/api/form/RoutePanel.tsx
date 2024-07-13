import React, {useState} from "react";
import {
    Button,
    Col,
    Collapse,
    Flex,
    Form,
    Input,
    InputNumber,
    Row,
    Select,
    Switch,
    Modal,
    Dropdown,
    Space,
    notification,
    type MenuProps
} from "antd";
import {CloseOutlined, DownOutlined} from "@ant-design/icons";
import {FormListFieldData} from "antd/es/form/FormList";
import {useTranslation} from "react-i18next";
import {RoutePanelData, Method, ContentType, GolangType} from "./_defaultProps";
import CodeMirror, {EditorView} from '@uiw/react-codemirror';
import {githubLight} from '@uiw/codemirror-theme-github';
import {langs} from '@uiw/codemirror-extensions-langs';
import type {FormInstance} from "antd/es/form/hooks/useForm";

const {TextArea} = Input;

interface RoutePanelProps {
    routeGroupField: FormListFieldData
    form: FormInstance
}


const RoutePanel: React.FC<RoutePanelProps & React.RefAttributes<HTMLDivElement>> = (props) => {
    const {t, i18n} = useTranslation();
    const routeGroupField = props.routeGroupField
    const form = props.form
    const [initRequestValues, setInitRequestValues] = useState([]);
    const [open, setOpen] = useState(false);
    const [requestBodyParseCode, setRequestBodyParseCode] = useState('');
    const [api, contextHolder] = notification.useNotification();
    const [showImportButton, setShowImportButton] = useState(true);

    const canChowImportButton = (routeIdx: number) => {
        const routeGroups = form.getFieldValue(`routeGroups`)
        if (!routeGroups) {
            setShowImportButton(true)
            return
        }

        if (routeGroups.length <= routeGroupField.key) {
            setShowImportButton(true)
            return
        }

        const routeGroup = routeGroups[routeGroupField.key]

        if (!routeGroup) {
            setShowImportButton(true)
            return
        }
        if (!routeGroup.routes) {
            setShowImportButton(true)
            return
        }

        if (routeGroup.routes.length <= routeIdx) {
            setShowImportButton(true)
            return
        }

        const route = routeGroup.routes[routeIdx]
        if (!route) {
            setShowImportButton(true)
            return
        }
        if (!route.requestBodyFields) {
            setShowImportButton(true)
            return
        }


        setShowImportButton(route.requestBodyFields.length === 0)
    }
    return (
        <div>
            {contextHolder}
            <Modal
                title={t("formRequestBodyFieldBtnImport")}
                centered
                open={open}
                maskClosable={false}
                keyboard={false}
                closable={false}
                destroyOnClose
                onOk={() => {
                    try {
                        const obj = JSON.parse(requestBodyParseCode)
                        if (Array.isArray(obj)) {
                            api.error({
                                message: t("tipsInvalidJSONArray")
                            })
                            return
                        }

                        // todo: 从后段解析数据
                        setOpen(false)
                    } catch (err) {
                        api.error({
                            message: t("tipsInvalidJSON") + ": " + err
                        })
                        return
                    }
                }}
                onCancel={() => setOpen(false)}
                width={1000}
                cancelText={t("formRequestBodyModalCancel")}
                okText={t("formRequestBodyModalConfirm")}
            >
                <CodeMirror
                    style={{marginTop: 10, overflow: "auto"}}
                    extensions={[langs.json(), EditorView.theme({
                        "&.cm-focused": {
                            outline: "none",
                        },
                    })]}
                    theme={githubLight}
                    height={'70vh'}
                    onChange={(code) => {
                        setRequestBodyParseCode(code)
                    }}
                />
            </Modal>
            <Form.Item label={t("formRouteListTitle")}>
                <Form.List
                    name={[routeGroupField.name, 'routes']}>
                    {(routeFields, routeOpt) => (
                        <div style={{
                            display: 'flex',
                            rowGap: 16,
                            flexDirection: 'column'
                        }}>

                            {routeFields.map((routeField) => (
                                <Collapse
                                    defaultActiveKey={[routeField.key]}
                                    items={[
                                        {
                                            key: routeField.key,
                                            label: t("formRouteTitle") + `${routeField.name + 1}`,
                                            children: <div>
                                                <Row gutter={16}>
                                                    <Col span={12}>
                                                        <Form.Item
                                                            label={t("formMethodTitle")}
                                                            name={[routeField.name, 'method']}>
                                                            <Select
                                                                defaultValue={Method.POST}
                                                                options={RoutePanelData.MethodOptions}
                                                            />
                                                        </Form.Item>
                                                    </Col>
                                                    <Col span={12}>
                                                        <Form.Item
                                                            label={t("formContentTypeTitle")}
                                                            name={[routeField.name, 'contentType']}>
                                                            <Select
                                                                defaultValue={ContentType.ApplicationJson}
                                                                options={RoutePanelData.ContentTypeOptions}
                                                            />
                                                        </Form.Item>
                                                    </Col>
                                                </Row>

                                                <Form.Item
                                                    label={t("formPathTitle")}
                                                    name={[routeField.name, 'path']}>
                                                    <Input/>
                                                </Form.Item>

                                                {/*request body*/}
                                                <Form.Item
                                                    label={t("formRequestBodyTitle")}>
                                                    <Form.List
                                                        initialValue={initRequestValues}
                                                        name={[routeField.name, 'requestBodyFields']}>
                                                        {(requestBodyFields, requestBodyOpt) => (
                                                            <div
                                                                style={{
                                                                    display: 'flex',
                                                                    flexDirection: 'column',
                                                                }}>

                                                                {requestBodyFields.map((requestBodyField) => (
                                                                    <Flex
                                                                        key={requestBodyField.key}
                                                                        gap={10}
                                                                    >
                                                                        <Form.Item
                                                                            label={t("formRequestBodyFieldNameTitle")}
                                                                            name={[requestBodyField.name, 'name']}
                                                                            style={{flex: 1}}
                                                                        >
                                                                            <Input/>
                                                                        </Form.Item>
                                                                        <Form.Item
                                                                            label={t("formRequestBodyFieldTypeTitle")}
                                                                            name={[requestBodyField.name, 'type']}
                                                                            style={{flex: 1}}
                                                                        >
                                                                            <Select
                                                                                defaultValue={GolangType.String}
                                                                                options={RoutePanelData.GolangTypeOptions}
                                                                            />
                                                                        </Form.Item>
                                                                        <CloseOutlined
                                                                            onClick={() => {
                                                                                requestBodyOpt.remove(requestBodyField.name);
                                                                                canChowImportButton(routeField.key)
                                                                            }}
                                                                        />
                                                                    </Flex>
                                                                ))}
                                                                {showImportButton ? <Button
                                                                    style={{marginBottom: 16}}
                                                                    type="dashed"
                                                                    onClick={() => setOpen(true)}
                                                                    block>
                                                                    🔍 {t("formRequestBodyFieldBtnImport")}
                                                                </Button> : <></>
                                                                }
                                                                <Button
                                                                    type="dashed"
                                                                    onClick={() => {
                                                                        requestBodyOpt.add()
                                                                        canChowImportButton(routeField.key)

                                                                    }}
                                                                    block>
                                                                    + {t("formRequestBodyFieldBtnAdd")}
                                                                </Button>

                                                            </div>

                                                        )}
                                                    </Form.List>
                                                </Form.Item>
                                                {/*  response body  */}
                                                <Form.Item
                                                    label={t("formResponseBodyTitle")}
                                                    name={[routeField.name, 'responseBody']}>
                                                    <TextArea
                                                        autoSize={{
                                                            minRows: 3,
                                                            maxRows: 5
                                                        }}
                                                        placeholder={t("formResponseBodyPlaceholder")}/>
                                                </Form.Item>
                                            </div>,
                                            extra: <CloseOutlined
                                                onClick={() => {
                                                    routeOpt.remove(routeField.name);
                                                }}
                                            />
                                        }
                                    ]}
                                />
                            ))}
                            <Button type="dashed"
                                    onClick={() => routeOpt.add()}
                                    block>
                                + {t("formButtonRouteAdd")}
                            </Button>

                        </div>

                    )}
                </Form.List>
            </Form.Item>
        </div>
    )
}

export default RoutePanel;