import { AGENT_STATUS, EXPRESSION_STATUS } from "./enums";

const greenIcon: string = "https://thumbs.dreamstime.com/b/%D0%B7%D0%B5%D0%BB%D0%B5%D0%BD%D0%B0%D1%8F-%D0%B3%D0%B0%D0%BB%D0%BE%D1%87%D0%BA%D0%B0-%D0%BF%D0%BE%D0%B4%D1%82%D0%B2%D0%B5%D1%80%D0%B6%D0%B4%D0%B0%D1%8E%D1%82-%D0%B8%D0%BB%D0%B8-%D0%BB%D0%B8%D0%BD%D0%B8%D1%8F-%D0%B7%D0%BD%D0%B0%D1%87%D0%BA%D0%B8-%D0%BA%D0%BE%D0%BD%D1%82%D1%80%D0%BE%D0%BB%D1%8C%D0%BD%D0%BE%D0%B9-%D0%BF%D0%BE%D0%BC%D0%B5%D1%82%D0%BA%D0%B8-185588596.jpg";
const yellowIcon: string = "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSeendsSC1KLb1ljM9hILUWPjJDcu_JPgdvsymh6DZj0YqBILgplXiIpWlELjhynu4R-9M&usqp=CAU";
const redIcon: string = "https://cdn-icons-png.flaticon.com/512/6368/6368418.png";
const blackIcon: string = "https://cdn-icons-png.flaticon.com/512/7351/7351882.png";

export const ICONS = {
  [AGENT_STATUS.RUNNING]: blackIcon,
  [AGENT_STATUS.SLEEPING]: greenIcon,
  [AGENT_STATUS.WAITING]: yellowIcon,
  [AGENT_STATUS.TERMINATED]: redIcon,
  [EXPRESSION_STATUS.READY_FOR_COMPUTATION]: blackIcon,
  [EXPRESSION_STATUS.RESULT]: greenIcon,
  [EXPRESSION_STATUS.COMPUTING]: yellowIcon,
} as const;

export const AGENT_DESCRIPTION = {
  [AGENT_STATUS.RUNNING]: "the server is calculating expressions and waiting for new ones",
  [AGENT_STATUS.SLEEPING]: "the server is calculating the expressions and is fully occupied",
  [AGENT_STATUS.WAITING]: "the server is waiting for new expressions",
  [AGENT_STATUS.TERMINATED]: "the server is down",
} as const;

export const EXPRESSION_DESCRIPTION = {
  [EXPRESSION_STATUS.READY_FOR_COMPUTATION]: "the expression is accepted, it will be processed soon",
  [EXPRESSION_STATUS.RESULT]: "the expression is ready",
  [EXPRESSION_STATUS.COMPUTING]: "the expression is being processed, it will be calculated soon",
  [EXPRESSION_STATUS.TERMINATED]: "expression parsing error",
} as const;